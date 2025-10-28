package handlers

import (
	"net/http"

	"claude-proxy/modules/auth/application/dto"
	"claude-proxy/modules/auth/domain/entities"
	"claude-proxy/modules/auth/domain/interfaces"
	"claude-proxy/pkg/errors"

	"github.com/gin-gonic/gin"
)

// AccountHandler handles HTTP requests for account management
type AccountHandler struct {
	accountService interfaces.AccountService
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(
	accountService interfaces.AccountService,
) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

// ListAccounts handles GET /api/accounts
func (h *AccountHandler) ListAccounts(c *gin.Context) {
	accounts, err := h.accountService.ListAccounts(c.Request.Context())
	if err != nil {
		panic(errors.NewInternalError("ACCOUNTS_LIST_FAILED", "Failed to list accounts", err.Error()))
	}

	accountResponses := make([]*dto.AccountResponse, len(accounts))
	for i, account := range accounts {
		accountResponses[i] = dto.ToAccountResponse(account)
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": accountResponses,
	})
}

// GetAccount handles GET /api/accounts/:id
func (h *AccountHandler) GetAccount(c *gin.Context) {
	id := c.Param("id")

	account, err := h.accountService.GetAccount(c.Request.Context(), id)
	if err != nil {
		panic(errors.NewNotFoundError("ACCOUNT_NOT_FOUND", "Account not found", id))
	}

	c.JSON(http.StatusOK, gin.H{
		"account": dto.ToAccountResponse(account),
	})
}

// UpdateAccount handles PUT /api/accounts/:id
func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		panic(errors.NewBadRequestError("INVALID_REQUEST", "Invalid request body", err.Error()))
	}

	// Update using service method
	var name string
	if req.Name != nil {
		name = *req.Name
	}
	var status entities.AccountStatus
	if req.Status != nil {
		status = entities.AccountStatus(*req.Status)
	}

	account, err := h.accountService.UpdateAccount(c.Request.Context(), id, name, status)
	if err != nil {
		panic(errors.NewInternalError("ACCOUNT_UPDATE_FAILED", "Failed to update account", err.Error()))
	}

	c.JSON(http.StatusOK, gin.H{
		"account": dto.ToAccountResponse(account),
	})
}

// DeleteAccount handles DELETE /api/accounts/:id
func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	id := c.Param("id")

	if err := h.accountService.DeleteAccount(c.Request.Context(), id); err != nil {
		panic(errors.NewNotFoundError("ACCOUNT_NOT_FOUND", "Account not found", id))
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "account deleted successfully",
	})
}
