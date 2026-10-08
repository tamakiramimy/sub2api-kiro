//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/suite"
)

type UsageLogKiroSessionSuite struct {
	suite.Suite
	ctx    context.Context
	tx     *dbent.Tx
	client *dbent.Client
	repo   *usageLogRepository
	user   *service.User
	apiKey *service.APIKey
}

func TestUsageLogKiroSessionSuite(t *testing.T) {
	suite.Run(t, new(UsageLogKiroSessionSuite))
}

func (s *UsageLogKiroSessionSuite) SetupTest() {
	s.ctx = context.Background()
	tx := testEntTx(s.T())
	s.tx = tx
	s.client = tx.Client()
	s.repo = newUsageLogRepositoryWithSQL(s.client, tx)
	s.user = mustCreateUser(s.T(), s.client, &service.User{Email: "kiro-session-" + uuid.New().String() + "@test.com"})
	s.apiKey = mustCreateApiKey(s.T(), s.client, &service.APIKey{UserID: s.user.ID, Key: "sk-" + uuid.New().String(), Name: "k"})
}

func (s *UsageLogKiroSessionSuite) createKiroLog(account *service.Account, fingerprint string, createdAt time.Time) *service.UsageLog {
	log := &service.UsageLog{
		UserID:                 s.user.ID,
		APIKeyID:               s.apiKey.ID,
		AccountID:              account.ID,
		RequestID:              uuid.New().String(),
		Model:                  "claude-sonnet-4-5",
		InputTokens:            10,
		OutputTokens:           5,
		TotalCost:              0.1,
		ActualCost:             0.1,
		KiroSessionFingerprint: &fingerprint,
		CreatedAt:              createdAt,
	}
	_, err := s.repo.Create(s.ctx, log)
	s.Require().NoError(err)
	return log
}

func (s *UsageLogKiroSessionSuite) listAll() []service.UsageLog {
	logs, _, err := s.repo.ListWithFilters(s.ctx, pagination.PaginationParams{Page: 1, PageSize: 50},
		usagestats.UsageLogFilters{UserID: s.user.ID, ExactTotal: true})
	s.Require().NoError(err)
	return logs
}

func (s *UsageLogKiroSessionSuite) findByRequestID(logs []service.UsageLog, requestID string) *service.UsageLog {
	for i := range logs {
		if logs[i].RequestID == requestID {
			return &logs[i]
		}
	}
	s.Require().Fail("request id not found", requestID)
	return nil
}

// TestFirstBindHasNoPreviousAccount 首次绑定的会话没有历史，previous 应为 nil。
func (s *UsageLogKiroSessionSuite) TestFirstBindHasNoPreviousAccount() {
	account := mustCreateAccount(s.T(), s.client, &service.Account{Name: "kiro-acc-first", Platform: service.PlatformKiro})
	fingerprint := "fp-" + uuid.New().String()
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	first := s.createKiroLog(account, fingerprint, base)

	logs := s.listAll()
	got := s.findByRequestID(logs, first.RequestID)
	s.Require().Nil(got.PreviousKiroAccountID)
}

// TestSameAccountReuseIsKept 同一账号在同一会话内重复出现时应标记为"保持"。
func (s *UsageLogKiroSessionSuite) TestSameAccountReuseIsKept() {
	account := mustCreateAccount(s.T(), s.client, &service.Account{Name: "kiro-acc-kept", Platform: service.PlatformKiro})
	fingerprint := "fp-" + uuid.New().String()
	base := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

	s.createKiroLog(account, fingerprint, base)
	second := s.createKiroLog(account, fingerprint, base.Add(time.Minute))

	logs := s.listAll()
	got := s.findByRequestID(logs, second.RequestID)
	s.Require().NotNil(got.PreviousKiroAccountID)
	s.Require().Equal(account.ID, *got.PreviousKiroAccountID)
	s.Require().Equal(account.ID, got.AccountID)
}

// TestAccountChangeIsSwitched 同一会话内更换账号时应能看到旧账号 ID。
func (s *UsageLogKiroSessionSuite) TestAccountChangeIsSwitched() {
	accountA := mustCreateAccount(s.T(), s.client, &service.Account{Name: "kiro-acc-switch-a", Platform: service.PlatformKiro})
	accountB := mustCreateAccount(s.T(), s.client, &service.Account{Name: "kiro-acc-switch-b", Platform: service.PlatformKiro})
	fingerprint := "fp-" + uuid.New().String()
	base := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)

	s.createKiroLog(accountA, fingerprint, base)
	switched := s.createKiroLog(accountB, fingerprint, base.Add(time.Minute))

	logs := s.listAll()
	got := s.findByRequestID(logs, switched.RequestID)
	s.Require().NotNil(got.PreviousKiroAccountID)
	s.Require().Equal(accountA.ID, *got.PreviousKiroAccountID)
	s.Require().Equal(accountB.ID, got.AccountID)
	s.Require().NotEqual(*got.PreviousKiroAccountID, got.AccountID)
}

// TestDifferentFingerprintsAreIsolated 不同会话指纹的账号历史互不影响。
func (s *UsageLogKiroSessionSuite) TestDifferentFingerprintsAreIsolated() {
	accountA := mustCreateAccount(s.T(), s.client, &service.Account{Name: "kiro-acc-iso-a", Platform: service.PlatformKiro})
	accountB := mustCreateAccount(s.T(), s.client, &service.Account{Name: "kiro-acc-iso-b", Platform: service.PlatformKiro})
	fingerprintA := "fp-a-" + uuid.New().String()
	fingerprintB := "fp-b-" + uuid.New().String()
	base := time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC)

	s.createKiroLog(accountA, fingerprintA, base)
	secondForA := s.createKiroLog(accountA, fingerprintA, base.Add(time.Minute))
	firstForB := s.createKiroLog(accountB, fingerprintB, base.Add(2*time.Minute))

	logs := s.listAll()
	gotSecondA := s.findByRequestID(logs, secondForA.RequestID)
	s.Require().NotNil(gotSecondA.PreviousKiroAccountID)
	s.Require().Equal(accountA.ID, *gotSecondA.PreviousKiroAccountID)

	gotFirstB := s.findByRequestID(logs, firstForB.RequestID)
	s.Require().Nil(gotFirstB.PreviousKiroAccountID, "a different fingerprint's own first entry must have no previous account")
}

// TestTieBreaksByIDWhenCreatedAtMatches 相同 created_at 时按 id 升序作为次要排序键。
func (s *UsageLogKiroSessionSuite) TestTieBreaksByIDWhenCreatedAtMatches() {
	accountA := mustCreateAccount(s.T(), s.client, &service.Account{Name: "kiro-acc-tie-a", Platform: service.PlatformKiro})
	accountB := mustCreateAccount(s.T(), s.client, &service.Account{Name: "kiro-acc-tie-b", Platform: service.PlatformKiro})
	fingerprint := "fp-" + uuid.New().String()
	same := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)

	first := s.createKiroLog(accountA, fingerprint, same)
	second := s.createKiroLog(accountB, fingerprint, same)
	s.Require().Less(first.ID, second.ID, "insertion order must produce increasing ids for the tie-break to be meaningful")

	logs := s.listAll()
	got := s.findByRequestID(logs, second.RequestID)
	s.Require().NotNil(got.PreviousKiroAccountID)
	s.Require().Equal(accountA.ID, *got.PreviousKiroAccountID)
}

// TestSurvivesDeletedAccount 账号被软删除后仍可回看历史切换记录。
func (s *UsageLogKiroSessionSuite) TestSurvivesDeletedAccount() {
	account := mustCreateAccount(s.T(), s.client, &service.Account{Name: "kiro-acc-deleted", Platform: service.PlatformKiro})
	fingerprint := "fp-" + uuid.New().String()
	base := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	s.createKiroLog(account, fingerprint, base)
	second := s.createKiroLog(account, fingerprint, base.Add(time.Minute))

	err := s.client.Account.UpdateOneID(account.ID).SetDeletedAt(time.Now().UTC()).Exec(s.ctx)
	s.Require().NoError(err, "soft delete account")

	logs := s.listAll()
	got := s.findByRequestID(logs, second.RequestID)
	s.Require().NotNil(got.PreviousKiroAccountID)
	s.Require().Equal(account.ID, *got.PreviousKiroAccountID)
}
