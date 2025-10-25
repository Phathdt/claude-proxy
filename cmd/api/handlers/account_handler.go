package handlers

import (
	"net/http"

	"claude-proxy/modules/proxy/application/dto"
	"claude-proxy/modules/proxy/domain/entities"
	"claude-proxy/modules/proxy/domain/interfaces"
	"claude-proxy/pkg/errors"

	"github.com/gin-gonic/gin"
)

// AccountHandler handles HTTP requests for account management
type AccountHandler struct {
	accountService interfaces.AccountService
	accountRepo    interfaces.AccountRepository
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(
	accountService interfaces.AccountService,
	accountRepo interfaces.AccountRepository,
) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
		accountRepo:    accountRepo,
	}
}

// ListAccounts handles GET /api/accounts
func (h *AccountHandler) ListAccounts(c *gin.Context) {
	accounts, err := h.accountRepo.List(c.Request.Context())
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

	account, err := h.accountRepo.GetByID(c.Request.Context(), id)
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

	account, err := h.accountRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		panic(errors.NewNotFoundError("ACCOUNT_NOT_FOUND", "Account not found", id))
	}

	// Update fields
	if req.Name != nil {
		account.Name = *req.Name
	}
	if req.Status != nil {
		account.Status = entities.AccountStatus(*req.Status)
	}

	if err := h.accountRepo.Update(c.Request.Context(), account); err != nil {
		panic(errors.NewInternalError("ACCOUNT_UPDATE_FAILED", "Failed to update account", err.Error()))
	}

	c.JSON(http.StatusOK, gin.H{
		"account": dto.ToAccountResponse(account),
	})
}

// DeleteAccount handles DELETE /api/accounts/:id
func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	id := c.Param("id")

	if err := h.accountRepo.Delete(c.Request.Context(), id); err != nil {
		panic(errors.NewNotFoundError("ACCOUNT_NOT_FOUND", "Account not found", id))
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "account deleted successfully",
	})
}
