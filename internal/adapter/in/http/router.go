package http

import "net/http"

type Router struct {
	Auth     *AuthHandler
	Room     *RoomHandler
	Schedule *ScheduleHandler
	Slot     *SlotHandler
}

func NewRouter(
	auth *AuthHandler,
	room *RoomHandler,
	schedule *ScheduleHandler,
	slot *SlotHandler,
) *Router {
	return &Router{
		Auth:     auth,
		Room:     room,
		Schedule: schedule,
		Slot:     slot,
	}
}

func (r *Router) InitRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /dummyLogin", r.Auth.DummyLogin)
	mux.HandleFunc("POST /register", r.Auth.Register)
	mux.HandleFunc("POST /login", r.Auth.Login)

	mux.HandleFunc("POST /rooms/create", r.Room.CreateRoom)
	mux.HandleFunc("GET /rooms/list", r.Room.ListRooms)

	mux.HandleFunc("POST /rooms/{roomId}/schedule/create", r.Schedule.CreateSchedule)

	mux.HandleFunc("GET /rooms/{roomId}/slots/list", r.Slot.ListSlots)

	var handler http.Handler = mux
	handler = r.withLogger(handler)
	handler = r.withRecovery(handler)

	return handler
}

func (r *Router) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, req)
	})
}

func (r *Router) withLogger(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		println(req.Method, req.URL.Path)
		nextHandler.ServeHTTP(w, req)
	})
}
