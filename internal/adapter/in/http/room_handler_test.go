package http_test

//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"log/slog"
//	"net/http"
//	"net/http/httptest"
//	"os"
//	"strings"
//	"testing"
//	adapterhttp "todo-api/internal/adapter/in/http"
//	"todo-api/internal/adapter/out/storage"
//	"todo-api/internal/app/dto"
//	"todo-api/internal/app/usecase"
//	"todo-api/internal/domain/entity"
//)
//
//func TestTH_Create(t *testing.T) {
//	store := storage.NewDataStorage()
//	uc := usecase.NewCreateTodoUC(store)
//	testLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
//	handler := adapterhttp.NewTodoHandler(
//		testLogger,
//		uc,
//		nil,
//		nil,
//		nil,
//		nil,
//	)
//
//	router := adapterhttp.NewRouter(handler)
//	mux := router.InitRoutes()
//
//	t.Run("Success", func(t *testing.T) {
//		reqBody := `{
//						"title": "Clean room №1",
//						"description": "use water"
//					}`
//		request := httptest.NewRequest("POST", "/todos", strings.NewReader(reqBody))
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusCreated {
//			t.Fatalf("expected status 201, got %v", recorder.Code)
//		}
//
//		var response dto.CreateTodoResponse
//		_ = json.NewDecoder(recorder.Body).Decode(&response)
//		if response.id == 0 {
//			t.Error("expected non-zero id in response")
//		}
//	})
//
//	t.Run("Error - Invalid request body", func(t *testing.T) {
//		reqBody := `{"title": "something"`
//		request := httptest.NewRequest("POST", "/todos", strings.NewReader(reqBody))
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusBadRequest {
//			t.Errorf("expected status 400, got %v", recorder.Code)
//		}
//	})
//
//	t.Run("Error - Empty title (validation)", func(t *testing.T) {
//		reqBody := `{
//						"description": "nothing here",
//						"completed": true
//					}`
//		request := httptest.NewRequest("POST", "/todos", strings.NewReader(reqBody))
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusBadRequest {
//			t.Errorf("expected status 400, got %v", recorder.Code)
//		}
//	})
//
//	t.Run("Error - Todo already exists (duplicate)", func(t *testing.T) {
//		reqBody := `{
//						"id": 75,
//						"title": "Learn Go"
//					}`
//		request := httptest.NewRequest("POST", "/todos", strings.NewReader(reqBody))
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		reqBody = `{
//						"id": 75,
//						"title": "Learn Python",
//						"completed": true
//					}`
//		request = httptest.NewRequest("POST", "/todos", strings.NewReader(reqBody))
//		recorder = httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusBadRequest {
//			t.Errorf("expected status 400, got %v", recorder.Code)
//		}
//	})
//}
//
//func TestTH_Get(t *testing.T) {
//	store := storage.NewDataStorage()
//
//	targetID := int64(10)
//	_ = store.CreateTodo(context.Background(), &entity.Todo{
//		id:          targetID,
//		Title:       "Learn math",
//		Description: "using ai tools, youtube videos",
//	})
//
//	guc := usecase.NewGetTodoUC(store)
//	testLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
//	handler := adapterhttp.NewTodoHandler(testLogger, nil, guc, nil, nil, nil)
//	router := adapterhttp.NewRouter(handler)
//	mux := router.InitRoutes()
//
//	t.Run("Success", func(t *testing.T) {
//		request := httptest.NewRequest("GET", fmt.Sprintf("/todos/%d", targetID), nil)
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusOK {
//			t.Fatalf("expected status 200, got %v", recorder.Code)
//		}
//
//		var response dto.GetTodoResponse
//		_ = json.NewDecoder(recorder.Body).Decode(&response)
//
//		if response.Todo.id != targetID {
//			t.Errorf("expected id %d, got %d", targetID, response.Todo.id)
//		}
//	})
//
//	t.Run("Error - Invalid id format (not int)", func(t *testing.T) {
//		request := httptest.NewRequest("GET", "/todos/^", nil)
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusBadRequest {
//			t.Errorf("expected status 400, got %v", recorder.Code)
//		}
//	})
//
//	t.Run("Error - Invalid id format (negative int)", func(t *testing.T) {
//		request := httptest.NewRequest("GET", "/todos/-1", nil)
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusBadRequest {
//			t.Errorf("expected status 400, got %v", recorder.Code)
//		}
//	})
//
//	t.Run("Error - Todo not found", func(t *testing.T) {
//		request := httptest.NewRequest("GET", "/todos/80", nil)
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusNotFound {
//			t.Errorf("expected status 404, got %v", recorder.Code)
//		}
//	})
//}
//
//func TestTH_GetList(t *testing.T) {
//	store := storage.NewDataStorage()
//
//	_ = store.CreateTodo(context.Background(), &entity.Todo{
//		Title:       "Learn math",
//		Description: "using ai tools, youtube videos",
//	})
//	_ = store.CreateTodo(context.Background(), &entity.Todo{
//		Title:       "Learn english",
//		Description: "using ai tools, youtube videos",
//	})
//	_ = store.CreateTodo(context.Background(), &entity.Todo{
//		Title:       "Learn JS",
//		Description: "using ai tools, youtube videos",
//	})
//	_ = store.CreateTodo(context.Background(), &entity.Todo{
//		Title:       "Learn C++",
//		Description: "using ai tools, youtube videos",
//	})
//
//	gluc := usecase.NewGetTodoListUC(store)
//	testLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
//	handler := adapterhttp.NewTodoHandler(testLogger, nil, nil, nil, nil, gluc)
//	router := adapterhttp.NewRouter(handler)
//	mux := router.InitRoutes()
//
//	t.Run("Success", func(t *testing.T) {
//		request := httptest.NewRequest("GET", "/todos", nil)
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusOK {
//			t.Fatalf("expected status 200, got %v", recorder.Code)
//		}
//
//		var response dto.GetTodoListResponse
//		_ = json.NewDecoder(recorder.Body).Decode(&response)
//
//		if len(response.Todos) != 4 {
//			t.Errorf("expected length 4, got %d", len(response.Todos))
//		}
//	})
//
//	t.Run("Success with parameters", func(t *testing.T) {
//		var (
//			testLimit      = 2
//			testOffset     = 2
//			testFirstTitle = "Learn JS"
//		)
//
//		request := httptest.NewRequest(
//			"GET",
//			fmt.Sprintf("/todos?limit=%d&offset=%d",
//				testLimit,
//				testOffset,
//			),
//			nil,
//		)
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusOK {
//			t.Fatalf("expected status 200, got %v", recorder.Code)
//		}
//
//		var response dto.GetTodoListResponse
//		_ = json.NewDecoder(recorder.Body).Decode(&response)
//
//		if len(response.Todos) != 2 {
//			t.Errorf("expected length 2, got %d", len(response.Todos))
//		}
//
//		if response.Todos[0].Title != testFirstTitle {
//			t.Errorf("expected title %s, got %s", testFirstTitle, response.Todos[0].Title)
//		}
//	})
//
//	t.Run("Error - Invalid limit (negative int)", func(t *testing.T) {
//		request := httptest.NewRequest("GET", "/todos?limit=-1", nil)
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusBadRequest {
//			t.Fatalf("expected status 400, got %v", recorder.Code)
//		}
//	})
//
//	t.Run("Error - Invalid offset (negative int)", func(t *testing.T) {
//		request := httptest.NewRequest("GET", "/todos?offset=-1", nil)
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusBadRequest {
//			t.Fatalf("expected status 400, got %v", recorder.Code)
//		}
//	})
//}
//
//func TestTH_Update(t *testing.T) {
//	store := storage.NewDataStorage()
//
//	var targetID = int64(10)
//	_ = store.CreateTodo(context.Background(), &entity.Todo{
//		id:          targetID,
//		Title:       "Learn math",
//		Description: "using ai tools, youtube videos",
//	})
//
//	uuc := usecase.NewUpdateTodoUC(store)
//	testLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
//	handler := adapterhttp.NewTodoHandler(testLogger, nil, nil, uuc, nil, nil)
//	router := adapterhttp.NewRouter(handler)
//	mux := router.InitRoutes()
//
//	t.Run("Success", func(t *testing.T) {
//		reqBody := `{
//						"title": "New Title"
//					}`
//		request := httptest.NewRequest("PUT", fmt.Sprintf("/todos/%d", targetID), strings.NewReader(reqBody))
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != 200 {
//			t.Fatalf("expected status 200, got %d", recorder.Code)
//		}
//
//		var response dto.UpdateTodoResponse
//		_ = json.NewDecoder(recorder.Body).Decode(&response)
//
//		if response.id != targetID {
//			t.Errorf("expected id %d, got %d", targetID, response.id)
//		}
//		if !response.Updated {
//			t.Error("expected updated = true, got false")
//		}
//	})
//
//	t.Run("Error - Invalid request body", func(t *testing.T) {
//		reqBody := `{title": "something"`
//		request := httptest.NewRequest("PUT", fmt.Sprintf("/todos/%d", targetID), strings.NewReader(reqBody))
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusBadRequest {
//			t.Errorf("expected status 400, got %v", recorder.Code)
//		}
//	})
//
//	t.Run("Error - Invalid id format (not int)", func(t *testing.T) {
//		reqBody := `{
//						"title": "New Title"
//					}`
//		request := httptest.NewRequest("PUT", "/todos/^", strings.NewReader(reqBody))
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusBadRequest {
//			t.Errorf("expected status 400, got %v", recorder.Code)
//		}
//	})
//
//	t.Run("Error - Invalid id format (negative int)", func(t *testing.T) {
//		reqBody := `{
//						"title": "New Title"
//					}`
//		request := httptest.NewRequest("PUT", "/todos/-1", strings.NewReader(reqBody))
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusBadRequest {
//			t.Errorf("expected status 400, got %v", recorder.Code)
//		}
//	})
//
//	t.Run("Error - Empty title (validation)", func(t *testing.T) {
//		reqBody := `{
//						"description": "nothing here",
//						"completed": true
//					}`
//		request := httptest.NewRequest("PUT", fmt.Sprintf("/todos/%d", targetID), strings.NewReader(reqBody))
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusBadRequest {
//			t.Errorf("expected status 400, got %v", recorder.Code)
//		}
//	})
//
//	t.Run("Error - Todo not found", func(t *testing.T) {
//		reqBody := `{
//						"title": "New Title"
//					}`
//		request := httptest.NewRequest("PUT", "/todos/50", strings.NewReader(reqBody))
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusNotFound {
//			t.Errorf("expected status 404, got %v", recorder.Code)
//		}
//	})
//}
//
//func TestTH_Delete(t *testing.T) {
//	store := storage.NewDataStorage()
//
//	var targetID = int64(10)
//	_ = store.CreateTodo(context.Background(), &entity.Todo{
//		id:          targetID,
//		Title:       "Learn math",
//		Description: "using ai tools, youtube videos",
//	})
//
//	duc := usecase.NewDeleteTodoUC(store)
//	testLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
//	handler := adapterhttp.NewTodoHandler(testLogger, nil, nil, nil, duc, nil)
//	router := adapterhttp.NewRouter(handler)
//	mux := router.InitRoutes()
//
//	t.Run("Success", func(t *testing.T) {
//		request := httptest.NewRequest("DELETE", fmt.Sprintf("/todos/%d", targetID), nil)
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != 200 {
//			t.Fatalf("expected status 200, got %d", recorder.Code)
//		}
//
//		var response dto.DeleteTodoResponse
//		_ = json.NewDecoder(recorder.Body).Decode(&response)
//
//		if response.id != targetID {
//			t.Errorf("expected id %d, got %d", targetID, response.id)
//		}
//		if !response.Deleted {
//			t.Error("expected deleted = true, got false")
//		}
//	})
//
//	t.Run("Error - Invalid id format (not int)", func(t *testing.T) {
//		request := httptest.NewRequest("DELETE", "/todos/^", nil)
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusBadRequest {
//			t.Errorf("expected status 400, got %v", recorder.Code)
//		}
//	})
//
//	t.Run("Error - Invalid id format (negative int)", func(t *testing.T) {
//		request := httptest.NewRequest("DELETE", "/todos/-1", nil)
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusBadRequest {
//			t.Errorf("expected status 400, got %v", recorder.Code)
//		}
//	})
//
//	t.Run("Error - Todo not found", func(t *testing.T) {
//		request := httptest.NewRequest("DELETE", "/todos/50", nil)
//		recorder := httptest.NewRecorder()
//		mux.ServeHTTP(recorder, request)
//
//		if recorder.Code != http.StatusNotFound {
//			t.Errorf("expected status 404, got %v", recorder.Code)
//		}
//	})
//}
