package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
	target := "/cafe?city=%s&count=%d"

	for city, _ := range cafeList {
		requests := []struct {
			count int
			want  int
		}{
			{0, 0},
			{1, 1},
			{2, 2},
			{100, min(100, len(cafeList[city]))},
		}
		for _, v := range requests {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf(target, city, v.count), nil)
			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			strings.TrimSpace(response.Body.String())
			if response.Body.String() == "" {
				assert.Equal(t, v.want, 0)
				continue
			}
			resCount := strings.Split(response.Body.String(), ",")
			assert.Equal(t, v.want, len(resCount))
		}
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	target := "/cafe?city=moscow&search=%s"

	requests := []struct {
		search    string // передаваемое значение search
		wantCount int    // ожидаемое количество кафе в ответе
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf(target, v.search), nil)
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)

		strings.TrimSpace(response.Body.String())
		cafe := strings.Split(response.Body.String(), ",")
		var count int
		for _, cafeName := range cafe {
			if strings.Contains(strings.ToLower(cafeName), strings.ToLower(v.search)) {
				count++
			}
		}
		assert.Equal(t, v.wantCount, count)
	}
}
