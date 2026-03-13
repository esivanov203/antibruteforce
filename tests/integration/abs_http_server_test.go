package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/esivanov203/antibruteforce/internal/httpapi"
	"github.com/esivanov203/antibruteforce/internal/service"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func httpURL() string {
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	u := fmt.Sprintf("http://%s:%s/", host, port)

	return u
}

func makeRequest(bj []byte, route string) (httpapi.ResponseBody, int, error) {
	body := httpapi.ResponseBody{}
	urlBase := httpURL()

	//nolint:noctx
	resp, err := http.Post(urlBase+route, "application/json", bytes.NewBuffer(bj))
	if err != nil {
		return body, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return httpapi.ResponseBody{}, 0, err
	}

	err = json.Unmarshal(respBody, &body)
	if err != nil {
		return body, resp.StatusCode, err
	}

	return body, resp.StatusCode, err
}

//nolint:funlen
func TestAuth(t *testing.T) {
	err := godotenv.Load("../../.env")
	require.NoError(t, err)

	request := httpapi.AuthRequest{
		Login:    "John",
		Password: "password",
		IP:       "10.10.10.10",
	}

	// прямой brute force
	limit, err := strconv.Atoi(os.Getenv("LIMIT_LOGIN"))
	require.NoError(t, err)

	for i := 0; i < limit; i++ {
		request.Password += strconv.Itoa(i)

		bj, err := json.Marshal(request)
		require.NoError(t, err)

		body, status, err := makeRequest(bj, "auth")

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
		require.Len(t, body.Errors, 0)

		require.True(t, body.OK)
	}

	request.Password = "Pass10"
	bj, err := json.Marshal(request)
	require.NoError(t, err)

	body, status, err := makeRequest(bj, "auth")

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, body.Errors, 0)

	require.False(t, body.OK)

	// Чистка бакета
	resetRequest := httpapi.ResetBucketRequest{
		Login: request.Login,
		IP:    request.IP,
	}
	bj, err = json.Marshal(resetRequest)
	require.NoError(t, err)

	body, status, err = makeRequest(bj, "bucket/reset")

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, body.Errors, 0)
	require.True(t, body.OK)

	// бакет сбросился
	bj, err = json.Marshal(request)
	require.NoError(t, err)
	body, status, err = makeRequest(bj, "auth")

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)

	require.True(t, body.OK)

	// проверку сброса бакета IP в интегр. тестах избыточна
	// поскольку код бакета один и тот же + проверяется unit-тестами

	// обратный brute force
	limit, err = strconv.Atoi(os.Getenv("LIMIT_PASSWORD"))
	require.NoError(t, err)

	request.Password = "popular"

	for i := 0; i < limit; i++ {
		request.Login += strconv.Itoa(i)
		bj, err := json.Marshal(request)
		require.NoError(t, err)

		body, status, err := makeRequest(bj, "auth")

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		require.True(t, body.OK)
	}

	request.Login = "Final"
	bj, err = json.Marshal(request)
	require.NoError(t, err)

	body, status, err = makeRequest(bj, "auth")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)

	require.False(t, body.OK)

	// ip bucket
	// при последовательных запросах
	// из-за большого лимита часть токенов успевает восстанавливаться
	// поэтому шлем параллельные http-запросы
	limit, err = strconv.Atoi(os.Getenv("LIMIT_IP"))
	require.NoError(t, err)
	request.IP = "172.10.10.116"

	wg := sync.WaitGroup{}

	for i := 0; i < limit; i++ {
		wg.Add(1)
		go func(request httpapi.AuthRequest, i int) {
			defer wg.Done()
			request.Login += strconv.Itoa(i)
			request.Password += strconv.Itoa(i)

			bj, err := json.Marshal(request)
			require.NoError(t, err)

			body, status, err := makeRequest(bj, "auth")

			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status)

			require.True(t, body.OK)
		}(request, i)
	}

	wg.Wait()
	bj, err = json.Marshal(request)
	require.NoError(t, err)

	body, status, err = makeRequest(bj, "auth")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)

	require.False(t, body.OK)

	// with error
	// здесь два теста с валидацией и ошибкой при выполнении
	// все ошибки тестируются в unit-тестах
	request.IP = "10.10."
	bj, err = json.Marshal(request)
	require.NoError(t, err)

	body, status, err = makeRequest(bj, "auth")

	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, status)
	require.Contains(t, body.Errors, service.ErrInvalidIP.Error())

	request.Password = ""
	bj, err = json.Marshal(request)
	require.NoError(t, err)

	body, status, err = makeRequest(bj, "auth")

	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, status)
	require.Contains(t, body.Errors, "password required")
}

func TestBlackWhiteList(t *testing.T) {
	blackAddRequest := httpapi.SubnetRequest{
		Subnet: "10.100.100.100/20",
	}
	bj, err := json.Marshal(blackAddRequest)
	require.NoError(t, err)

	_, status, err := makeRequest(bj, "blacklist")
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, status)

	request := httpapi.AuthRequest{
		Login:    "Bob",
		Password: "bobpwd",
		IP:       "10.100.100.100",
	}
	bj, err = json.Marshal(request)
	require.NoError(t, err)
	body, status, err := makeRequest(bj, "auth")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.False(t, body.OK)

	client := &http.Client{}
	urlBase := httpURL()

	deleteURL := fmt.Sprintf("%sblacklist?subnet=%s", urlBase, blackAddRequest.Subnet)
	fmt.Println(deleteURL)
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodDelete,
		deleteURL,
		bytes.NewBuffer([]byte("{}")),
	)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	whiteAddRequest := httpapi.SubnetRequest{
		Subnet: "10.200.200.200/21",
	}
	bj, err = json.Marshal(whiteAddRequest)
	require.NoError(t, err)

	_, status, err = makeRequest(bj, "whitelist")
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, status)
}
