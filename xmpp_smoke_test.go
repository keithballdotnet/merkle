// +build smoke

package functionaltests

import (
	"brabbler/proto/mam2"
	"brabbler/tests/helpers/hosts"
	"brabbler/tests/helpers/mam"
	testxmpp "brabbler/tests/helpers/xmpp"
	"brabbler/util/authhelpers"
	"brabbler/util/client"
	"brabbler/util/config"
	"brabbler/util/gohelpers"
	"brabbler/util/shared"
	"brabbler/util/testutil"
	"brabbler/util/testutil/tabledriven"
	"brabbler/util/xmpp"
	"brabbler/util/xmpphelpers/xmppginlo"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func TestXMPPSmoke(t *testing.T) {
	config.Get().XMPP.ControlsVersion = xmppginlo.Version_1_1
	talk, access, credentials, err := testxmpp.InitOnce(t)
	require.NoError(t, err)

	testutil.TimeLimit(t, "Archive", 20*time.Second, func(t *testing.T) {
		// send three messages with some deay between them
		text := []string{"test 0", "':\"\\\"\\'\\:bla", "test 2"}
		conversation := "conv:" + uuid.NewV4().String()
		controls := []xmppginlo.XMPPControls{
			{MessageID: "msg:test:1:" + uuid.NewV4().String(), Ack: false, XMLName: xmppginlo.ElementName, ConversationID: conversation},
			{MessageID: "msg:test:2:" + uuid.NewV4().String(), Ack: true, NoArchive: true, ConversationID: conversation},
			{MessageID: "msg:test:3:" + uuid.NewV4().String(), Ack: true, ConversationID: conversation},
			{MessageID: "msg:test:4:" + uuid.NewV4().String(), Ack: false, ConversationID: "gid:wrong"},
		}
		xmppUsername := testxmpp.JID(&credentials)
		t.Logf("Sending user is '%s'", xmppUsername)
		require.NotEmpty(t, xmppUsername)
		start := time.Now().UTC()
		time.Sleep(time.Second)
		_, err = talk.Send(xmpp.Chat{Remote: xmppUsername, Type: "chat", Text: text[0], Controls: controls[0]})
		t.Logf("Sent message '%s' omit from archive = %v", controls[0].MessageID, controls[0].NoArchive)
		require.NoError(t, err)
		time.Sleep(time.Second)
		_, err = talk.Send(xmpp.Chat{Remote: xmppUsername, Type: "chat", Text: "should not be archived", Controls: controls[1]})
		t.Logf("Sent message '%s' omit from archive = %v", controls[1].MessageID, controls[1].NoArchive)
		_, err = talk.Send(xmpp.Chat{Remote: xmppUsername, Type: "chat", Text: text[1], Controls: controls[2]})
		t.Logf("Sent message '%s' omit from archive = %v", controls[2].MessageID, controls[2].NoArchive)
		require.NoError(t, err)
		time.Sleep(time.Second)
		_, err = talk.Send(xmpp.Chat{Remote: xmppUsername, Type: "chat", Text: text[2], Controls: controls[3]})
		t.Logf("Sent message '%s' omit from archive = %v", controls[3].MessageID, controls[3].NoArchive)
		require.NoError(t, err)
		time.Sleep(time.Second)
		end := time.Now().UTC()

		jwtCredentials := []grpc.DialOption{grpc.WithInsecure(), grpc.WithUserAgent("FunctionalTest"), authhelpers.NewTokenCredentials(access.AccessTokenCode)}
		host := hosts.API()
		grpchost := hosts.GRPC()
		index := strings.Index(host, ":")
		host = host[:index] + ":30083"

		t.Run("v2", func(t *testing.T) {
			getMinikubeSaveRemoteTimes := func(messages []mam2.Message) []time.Time {
				var timestamps []int64
				for _, msg := range messages {
					timestamps = append(timestamps, msg.Timestamp)
				}
				return gohelpers.SortedFromUnixnano(timestamps)
			}

			// make a backing off query to wait for DB to see all MAM entries
			veryLargeQuery := &mam2.Query{
				With:         credentials.UserID,
				Conversation: conversation,
				Start:        start.Add(-9999 * time.Hour).UnixNano(),
				End:          end.Add(9999 * time.Hour).UnixNano(),
			}
			allMessages := mam.Check(t, veryLargeQuery, access, 2)
			times := getMinikubeSaveRemoteTimes(allMessages)
			require.Len(t, times, 2)
			require.Equal(t, controls[0].MessageID, allMessages[1].Id) // MAM now sorts DESC => order inversed
			require.Equal(t, controls[2].MessageID, allMessages[0].Id)

			// now that we know the DB has the messages query MAM without back-off:
			testData := []struct {
				Query         *mam2.Query
				ExpectedError bool
				ExpectedCount int
				tabledriven.Description
			}{
				{&mam2.Query{}, false, 0, tabledriven.Describe("empty query yields nothing as empty conversation does not exist")},
				{&mam2.Query{With: credentials.UserID, Start: start.UTC().Add(-9999 * time.Hour).UnixNano(), End: end.UTC().Add(9999 * time.Hour).UnixNano(), Conversation: conversation}, false, 2, tabledriven.Describe("query with enlarged timeframe yields 2")},
				{&mam2.Query{With: credentials.UserID, Start: times[0].UnixNano(), End: times[1].UnixNano(), Conversation: conversation}, false, 2, tabledriven.Describe("query in timeframe yields 2")},
				{&mam2.Query{With: credentials.UserID, End: time.Now().UTC().Add(time.Hour).UnixNano(), Conversation: conversation}, false, 2, tabledriven.Describe("query yields all two")},
				{&mam2.Query{With: credentials.UserID, Start: times[0].UnixNano(), End: times[1].UnixNano(), Conversation: conversation}, false, 2, tabledriven.Describe("query since first archive yields 2")},
				{&mam2.Query{With: credentials.UserID, Start: times[1].UnixNano(), End: times[1].UnixNano(), Conversation: conversation}, false, 1, tabledriven.Describe("query since last archive yields 1")},
				{&mam2.Query{With: credentials.UserID, Start: times[0].UnixNano(), End: times[1].UnixNano(), Conversation: "wrong"}, false, 0, tabledriven.Describe("query for wrong conversation in timeframe yields 0")},
			}

			for _, testCase := range testData {
				t.Run("grpc/"+testCase.Description.Name(), func(t *testing.T) {
					// No secure connection?  Then run the grpc tests...
					if os.Getenv("USE_SSL") == "true" {
						t.Skipf("GRPC end-to-end tests currently disabled due to missing gateway")
					}
					_, errGet := mam.Get(t, grpchost, testCase.Query, testCase.ExpectedCount, jwtCredentials...)
					if testCase.ExpectedError {
						require.Error(t, errGet, tabledriven.Summary(testCase))
					} else {
						require.NoError(t, errGet, tabledriven.Summary(testCase))
					}
				})
				t.Run("json/"+testCase.Description.Name(), func(t *testing.T) {
					_, errGet := mam.GetViaJSON(testCase.Query, access, testCase.ExpectedCount)
					if testCase.ExpectedError {
						require.Error(t, errGet, tabledriven.Summary(testCase))
					} else {
						require.NoError(t, errGet, tabledriven.Summary(testCase))
					}

					t.Run("count", func(t *testing.T) {
						_, errCount := mam.CountViaJSON(shared.NewMAMCountQuery(testCase.Query), access, testCase.ExpectedCount)
						if testCase.ExpectedError {
							require.Error(t, errCount, tabledriven.Summary(testCase))
						} else {
							require.NoError(t, errCount, tabledriven.Summary(testCase))
						}
					})
				})
			}

			deleteViaJSON := func(query *mam2.DeleteRequest) error {
				jsonQuery := shared.MAMDelete{
					Mode:         query.Mode,
					With:         query.With,
					Timestamp:    strconv.FormatInt(query.Timestamp, 10),
					Conversation: query.Conversation,
					ID:           query.Id,
				}
				return client.DefaultClient.DeleteMessages(context.Background(), jsonQuery, access.AccessTokenCode)
			}

			// finally delete some messages
			t.Run("delete/second_by_ID", func(t *testing.T) {
				deleteViaJSON(&mam2.DeleteRequest{
					Mode:         shared.MAMByID,
					With:         allMessages[1].Sender,
					Timestamp:    allMessages[1].Timestamp,
					Conversation: allMessages[1].Conversation,
					Id:           allMessages[1].Id,
				})
				msgs, errGet := mam.GetViaJSON(veryLargeQuery, access, 1)
				require.NoError(t, errGet)
				require.NotContains(t, msgs, allMessages[1])
			})

			t.Run("delete/all_of_sender", func(t *testing.T) {
				deleteViaJSON(&mam2.DeleteRequest{
					Mode: shared.MAMByUser,
					With: credentials.UserID,
				})
				_, errGet := mam.GetViaJSON(veryLargeQuery, access, 0)
				require.NoError(t, errGet)
			})
		})
	})

	testutil.TimeLimit(t, "Message", 10*time.Second, func(t *testing.T) {
		text := fmt.Sprintf("test message %s", uuid.NewV4().String())
		xmppUsername := testxmpp.JID(&credentials)
		require.NotEmpty(t, xmppUsername)
		var msgID = "msg:test:" + uuid.NewV4().String()
		t.Logf("Sending user is '%s'", xmppUsername)
		_, err = talk.Send(xmpp.Chat{Remote: xmppUsername, Type: "chat", Text: text, Controls: xmppginlo.XMPPControls{MessageID: msgID}})
		require.NoError(t, err)
		t.Logf("Sent message '%s'", msgID)

		require.NoError(t, receive(talk, func(stanza interface{}) error {
			switch v := stanza.(type) {
			case xmpp.Chat:
				if v.Text == text {
					return nil
				}
				return fmt.Errorf("Unexpected text '%s' (wanted '%s')", v.Text, text)
			}
			return errors.New("Unexpected stanza type")
		}, 5*time.Second))
	})
}

func receive(talk *xmpp.Client, evalStanza func(stanza interface{}) error, timeout time.Duration) error {
	start := time.Now()
	done := make(chan error)
	go func() {
		for {
			if time.Since(start) > timeout {
				// no matter this will leak if Recv() never returns!
				return
			}
			chat, err := talk.Recv()
			if err != nil {
				done <- err
				return
			}
			err = evalStanza(chat)
			if err == nil {
				close(done)
				return
			}
		}
	}()

	select {
	case <-time.After(timeout):
		break
	case err := <-done:
		return err
	}

	return errors.New("Timeout")
}
