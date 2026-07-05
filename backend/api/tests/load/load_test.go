package load

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ──────────────────────────────────────────────────────────────
// Load tests
// ──────────────────────────────────────────────────────────────

// TestLoadSlotsList — самый высоконагруженный эндпоинт согласно ТЗ.
// SLI: 99.9% успешных ответов, p99 <= 200ms при 100 RPS.
func TestLoadSlotsList(t *testing.T) {
	a := setupLoad(t)

	fn := func() (*http.Response, error) {
		roomID := a.roomIDs[rand.Intn(len(a.roomIDs))]
		dayOffset := rand.Intn(30) + 1
		date := time.Now().Add(time.Duration(dayOffset) * 24 * time.Hour).Format(time.DateOnly)
		user := a.userPool[rand.Intn(len(a.userPool))]

		path := fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomID, date)
		return a.makeRequestAuth(http.MethodGet, path, nil, user.token)
	}

	results := runLoad(targetRPS, 100, 60*time.Second, fn)
	checkSLI(t, "GET /rooms/{id}/slots/list", results, slotsListP99Target)
}

func TestLoadBookingsCreate(t *testing.T) {
	a := setupLoad(t)

	require.NotEmpty(t, a.freeSlots, "no free slots reserved for booking load test")

	fn := func() (*http.Response, error) {
		select {
		case slotID := <-a.freeSlots:
			user := a.userPool[rand.Intn(len(a.userPool))]
			payload := map[string]interface{}{"slot_id": slotID, "create_conference_link": false}
			return a.makeRequestAuth(http.MethodPost, "/bookings/create", payload, user.token)
		default:
			return nil, fmt.Errorf("no free slots left")
		}
	}

	results := runLoad(targetRPS, 50, 30*time.Second, fn)
	checkSLI(t, "POST /bookings/create", results, 0)
}
