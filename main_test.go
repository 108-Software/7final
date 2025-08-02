package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		name     string
		count    int
		city     string
		want     int
		wantCode int
	}{
		{"count=0", 0, "moscow", 0, http.StatusOK},
		{"count=1", 1, "moscow", 1, http.StatusOK},
		{"count=2", 2, "moscow", 2, http.StatusOK},
		{"count=100", 100, "moscow", min(len(cafeList["moscow"]), 100), http.StatusOK},
	}

	for _, tt := range requests {
		t.Run(tt.name, func(t *testing.T) {

			url := "/cafe?count=" + strconv.Itoa(tt.count) + "&city=" + tt.city
			req := httptest.NewRequest("GET", url, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, req)
			require.Equal(t, tt.wantCode, response.Code)

			answer := strings.TrimSpace(response.Body.String())

			if answer == "" {
				require.Equal(t, 0, tt.count, "Пустой ответ возможен только при count=0")
				return
			}

			cafes := strings.Split(answer, ",")
			require.Len(t, cafes, tt.want)
		})
	}
}

func TestCafeSearch(t *testing.T) {

	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		name      string
		search    string 
		wantCount int    
	}{
		{"search=фасоль", "фасоль", 0}, //Я захотел расширить поиск
		{"search=мир", "мир", 1},
		{"search=сладкоежка", "сладкоежка", 1},
		{"search=завтраки", "завтраки", 1},
		{"search=студент", "студент", 1},
		{"search=сытый", "сытый", 1},
		{"search=ложка", "ложка", 1},
		{"search=кофе", "кофе", 2},
		{"search=вилка", "вилка", 1},
	}

	for _, tt := range requests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/cafe?city=moscow&search=" + tt.search
			req := httptest.NewRequest("GET", url, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			body := response.Body.String()
			var cafes []string
			if body != "" {
				cafes = strings.Split(body, ",")
				for i := range cafes {
					cafes[i] = strings.TrimSpace(cafes[i])
				}
			}

			assert.Len(t, cafes, tt.wantCount, "Неверное количество кафе в ответе")

			if tt.wantCount > 0 {
				searchLower := strings.ToLower(tt.search)
				for _, cafe := range cafes {
					cafeLower := strings.ToLower(cafe)
					assert.Contains(t, cafeLower, searchLower,
						"Кафе '%s' должно содержать подстроку '%s'", cafe, tt.search)
				}
			}
		})
	}

}
