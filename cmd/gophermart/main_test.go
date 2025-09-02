package main

import (
	"billBox/internal/api"
	"billBox/internal/config"
	"billBox/internal/models"
	"billBox/internal/storage/mocks"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGophermartFlowWithMockStorage(t *testing.T) {
	// --- 0. Настройка ---
	mockStorage := new(mocks.Storage)
	logger := log.New(os.Stdout, "test-logger ", log.LstdFlags)
	cfg := &config.Config{JWTSecretKey: "very-secret-key-for-tests"}
	router := api.NewRouter(mockStorage, logger, cfg)

	// --- Шаг 1: Регистрация пользователя ---
	userLogin := "testuser"
	userPassword := "testpassword123"
	var registeredUserID int64 = 1 // ID, который мы ожидаем получить после регистрации

	// Настраиваем мок для CreateUser
	mockStorage.On("CreateUser", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.Login == userLogin
	})).Return(nil).Run(func(args mock.Arguments) {
		user := args.Get(1).(*models.User)
		user.ID = registeredUserID // Симулируем присвоение ID
	}).Once()

	regBody, _ := json.Marshal(models.AuthRequest{Login: userLogin, Password: userPassword})
	regReq := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(regBody))
	regRR := httptest.NewRecorder()
	router.ServeHTTP(regRR, regReq)

	require.Equal(t, http.StatusOK, regRR.Code, "Регистрация должна пройти успешно")

	// Получаем cookie для следующих запросов
	res := regRR.Result()
	defer res.Body.Close()
	authCookie := res.Cookies()[0]
	require.NotNil(t, authCookie)

	// --- Шаг 2: Загрузка номера заказа ---
	orderNumber := "12345678903"
	mockStorage.On("FindByNumber", mock.Anything, orderNumber).Return(nil, nil).Once()
	mockStorage.On("CreateOrder", mock.Anything, mock.MatchedBy(func(order *models.Order) bool {
		return order.Number == orderNumber && order.UserID == registeredUserID
	})).Return(nil).Once()

	uploadReq := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(orderNumber))
	uploadReq.AddCookie(authCookie)
	uploadRR := httptest.NewRecorder()
	router.ServeHTTP(uploadRR, uploadReq)

	require.Equal(t, http.StatusAccepted, uploadRR.Code, "Новый заказ должен быть принят")

	// --- Шаг 3: Проверка баланса (симулируем, что начислено 500 баллов) ---
	mockStorage.On("Get", mock.Anything, registeredUserID).Return(&models.Balance{Current: 500, Withdrawn: 0}, nil).Once()

	balanceReq1 := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	balanceReq1.AddCookie(authCookie)
	balanceRR1 := httptest.NewRecorder()
	router.ServeHTTP(balanceRR1, balanceReq1)

	require.Equal(t, http.StatusOK, balanceRR1.Code)
	var balance1 models.Balance
	err := json.NewDecoder(balanceRR1.Body).Decode(&balance1)
	require.NoError(t, err)
	assert.Equal(t, 500.0, balance1.Current, "Текущий баланс должен быть 500")
	assert.Equal(t, 0.0, balance1.Withdrawn, "Списаний еще не было")

	// --- Шаг 4: Списание баллов ---
	withdrawAmount := 150.5
	// Номер заказа должен быть валидным по алгоритму Луна.
	withdrawOrder := "662315830601"
	mockStorage.On("Withdraw", mock.Anything, registeredUserID, withdrawOrder, withdrawAmount).Return(nil).Once()

	withdrawBody, _ := json.Marshal(models.WithdrawRequest{OrderNumber: withdrawOrder, Sum: withdrawAmount})
	withdrawReq := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(withdrawBody))
	withdrawReq.AddCookie(authCookie)
	withdrawRR := httptest.NewRecorder()
	router.ServeHTTP(withdrawRR, withdrawReq)

	require.Equal(t, http.StatusOK, withdrawRR.Code, "Списание должно пройти успешно")

	// --- Шаг 5: Проверка баланса после списания ---
	mockStorage.On("Get", mock.Anything, registeredUserID).Return(&models.Balance{Current: 349.5, Withdrawn: 150.5}, nil).Once()

	balanceReq2 := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	balanceReq2.AddCookie(authCookie)
	balanceRR2 := httptest.NewRecorder()
	router.ServeHTTP(balanceRR2, balanceReq2)

	require.Equal(t, http.StatusOK, balanceRR2.Code)
	var balance2 models.Balance
	err = json.NewDecoder(balanceRR2.Body).Decode(&balance2)
	require.NoError(t, err)
	assert.Equal(t, 349.5, balance2.Current, "Текущий баланс должен уменьшиться")
	assert.Equal(t, 150.5, balance2.Withdrawn, "Сумма списаний должна увеличиться")

	// --- Шаг 6: Проверка истории списаний ---
	expectedWithdrawals := []models.Withdrawal{
		{OrderNumber: withdrawOrder, Sum: withdrawAmount},
	}
	mockStorage.On("WithdrawalFindForUser", mock.Anything, registeredUserID).Return(expectedWithdrawals, nil).Once()

	historyReq := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	historyReq.AddCookie(authCookie)
	historyRR := httptest.NewRecorder()
	router.ServeHTTP(historyRR, historyReq)

	require.Equal(t, http.StatusOK, historyRR.Code)
	var withdrawals []models.Withdrawal
	err = json.NewDecoder(historyRR.Body).Decode(&withdrawals)
	require.NoError(t, err)
	require.Len(t, withdrawals, 1, "В истории должно быть одно списание")
	assert.Equal(t, withdrawOrder, withdrawals[0].OrderNumber)
	assert.Equal(t, withdrawAmount, withdrawals[0].Sum)

	// --- Финальная проверка ---
	mockStorage.AssertExpectations(t)
}