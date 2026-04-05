package http

import (
	"MeetingRoomsAPI/internal/infrastructure/jwt"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

type Router struct {
	Auth     *AuthHandler
	Room     *RoomHandler
	Schedule *ScheduleHandler
	Slot     *SlotHandler
	Booking  *BookingHandler
	jwtGen   *jwt.TokenGenerator
}

func NewRouter(
	auth *AuthHandler,
	room *RoomHandler,
	schedule *ScheduleHandler,
	slot *SlotHandler,
	booking *BookingHandler,
	jwtGen *jwt.TokenGenerator,
) *Router {
	return &Router{
		Auth:     auth,
		Room:     room,
		Schedule: schedule,
		Slot:     slot,
		Booking:  booking,
		jwtGen:   jwtGen,
	}
}

func (r *Router) InitRoutes(log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("POST /dummyLogin", r.Auth.DummyLogin)
	mux.HandleFunc("POST /register", r.Auth.Register)
	mux.HandleFunc("POST /login", r.Auth.Login)

	// Private routes (user | admin)
	mux.Handle("GET /rooms/list", r.withAuth(http.HandlerFunc(r.Room.ListRooms)))
	mux.Handle("POST /rooms/create", r.withAuth(r.withRole("admin", http.HandlerFunc(r.Room.CreateRoom))))

	mux.Handle("POST /rooms/{roomId}/schedule/create", r.withAuth(r.withRole("admin", http.HandlerFunc(r.Schedule.CreateSchedule))))

	mux.Handle("GET /rooms/{roomId}/slots/list", r.withAuth(http.HandlerFunc(r.Slot.ListSlots)))

	mux.Handle("POST /bookings/create", r.withAuth(r.withRole("user", http.HandlerFunc(r.Booking.CreateBooking))))
	mux.Handle("GET /bookings/list", r.withAuth(r.withRole("admin", http.HandlerFunc(r.Booking.ListAllBookings))))
	mux.Handle("GET /bookings/my", r.withAuth(r.withRole("user", http.HandlerFunc(r.Booking.ListMyBookings))))
	mux.Handle("POST /bookings/{bookingId}/cancel", r.withAuth(r.withRole("user", http.HandlerFunc(r.Booking.CancelBooking))))

	var handler http.Handler = mux
	handler = r.withLogger(log, handler)
	handler = r.withRecovery(handler)

	return handler
}

// --- Middlewares ---

func (r *Router) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "unauthorized: missing token", http.StatusUnauthorized)
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "unauthorized: invalid auth format", http.StatusUnauthorized)
			return
		}

		userID, role, err := r.jwtGen.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, fmt.Sprintf("unauthorized: %v", err), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(req.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, RoleKey, role)

		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func (r *Router) withRole(requiredRole string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		role, ok := req.Context().Value(RoleKey).(string)
		if !ok || role != requiredRole {
			http.Error(w, "forbidden: insufficient permissions", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func (r *Router) withLogger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, req)

		log.InfoContext(req.Context(), "http request",
			slog.String("method", req.Method),
			slog.String("path", req.URL.Path),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

func (r *Router) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, req)
	})
}
