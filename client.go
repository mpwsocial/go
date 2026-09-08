package moipayway

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	TestBase = "https://dev.moipayway.co"
	LiveBase = "https://api.moipayway.co"
)

type Client struct {
	APIKey         string
	Environment    string
	HTTP           *http.Client
	Account        *Account
	Authentication *Authentication
	Card           *Card
	Misc           *Misc
	Omnichain      *Omnichain
	Simulation     *Simulation
	User           *User
	Verification   *Verification
	Wallet         *Wallet
}

type APIError struct {
	Message    string
	HTTPStatus int
	Payload    map[string]any
}

func (e *APIError) Error() string {
	return e.Message
}

func New(apiKey, environment string) (*Client, error) {
	env := strings.ToLower(strings.TrimSpace(environment))
	if env == "" {
		env = "test"
	}
	if env != "test" && env != "live" {
		return nil, fmt.Errorf("environment must be test or live")
	}
	c := &Client{
		APIKey:      strings.TrimSpace(apiKey),
		Environment: env,
		HTTP:        &http.Client{Timeout: 30 * time.Second},
	}
	c.Account = newAccount(c)
	c.Authentication = newAuthentication(c)
	c.Card = newCard(c)
	c.Misc = newMisc(c)
	c.Omnichain = newOmnichain(c)
	c.Simulation = newSimulation(c)
	c.User = newUser(c)
	c.Verification = newVerification(c)
	c.Wallet = newWallet(c)
	return c, nil
}

func (c *Client) BaseURL() string {
	if c.Environment == "live" {
		return LiveBase
	}
	return TestBase
}

func (c *Client) Request(method, path string, body map[string]any, auth bool) (map[string]any, error) {
	url := strings.TrimRight(c.BaseURL(), "/") + "/" + strings.TrimLeft(path, "/")
	var reader io.Reader
	if !strings.EqualFold(method, http.MethodGet) && !strings.EqualFold(method, http.MethodHead) {
		if body == nil {
			body = map[string]any{}
		}
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(strings.ToUpper(method), url, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if auth {
		if strings.TrimSpace(c.APIKey) == "" {
			return nil, fmt.Errorf("apiKey is required")
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var decoded map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &decoded); err != nil {
			return nil, &APIError{Message: "Invalid JSON response", HTTPStatus: res.StatusCode}
		}
	}
	if decoded == nil {
		decoded = map[string]any{}
	}
	status, _ := decoded["status"].(string)
	status = strings.ToLower(status)
	if status == "failed" || status == "error" {
		msg, _ := decoded["message"].(string)
		if msg == "" {
			msg = "Request failed"
		}
		return nil, &APIError{Message: msg, HTTPStatus: res.StatusCode, Payload: decoded}
	}
	return decoded, nil
}

func VerifyWebhook(rawBody, signatureHeader, timestampHeader, apiKey string, toleranceSeconds int) bool {
	if rawBody == "" || signatureHeader == "" || timestampHeader == "" || apiKey == "" {
		return false
	}
	ts, err := strconv.ParseInt(timestampHeader, 10, 64)
	if err != nil {
		return false
	}
	if toleranceSeconds > 0 {
		age := time.Now().Unix() - ts
		if age < 0 {
			age = -age
		}
		if age > int64(toleranceSeconds) {
			return false
		}
	}
	provided := signatureHeader
	if strings.HasPrefix(strings.ToLower(provided), "sha256=") {
		provided = provided[7:]
	}
	mac := hmac.New(sha256.New, []byte(apiKey))
	_, _ = mac.Write([]byte(timestampHeader + "." + rawBody))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(provided))
}

type Account struct {
	c        *Client
	Merchant *AccountMerchant
}

type AccountMerchant struct {
	c       *Client
	Product *AccountMerchantProduct
	Team    *AccountMerchantTeam
}

type AccountMerchantProduct struct {
	c            *Client
	Subscription *AccountMerchantProductSubscription
}

type AccountMerchantProductSubscription struct {
	c *Client
}

type AccountMerchantTeam struct {
	c          *Client
	Permission *AccountMerchantTeamPermission
}

type AccountMerchantTeamPermission struct {
	c *Client
}

type Authentication struct {
	c *Client
}

type Card struct {
	c *Client
}

type Misc struct {
	c *Client
}

type Omnichain struct {
	c            *Client
	Storage      *OmnichainStorage
	Tokenization *OmnichainTokenization
	Wallet       *OmnichainWallet
}

type OmnichainStorage struct {
	c *Client
}

type OmnichainTokenization struct {
	c       *Client
	Erc1155 *OmnichainTokenizationErc1155
	Erc20   *OmnichainTokenizationErc20
	Erc721  *OmnichainTokenizationErc721
	Sep41   *OmnichainTokenizationSep41
}

type OmnichainTokenizationErc1155 struct {
	c *Client
}

type OmnichainTokenizationErc20 struct {
	c *Client
}

type OmnichainTokenizationErc721 struct {
	c *Client
}

type OmnichainTokenizationSep41 struct {
	c *Client
}

type OmnichainWallet struct {
	c       *Client
	Evm     *OmnichainWalletEvm
	Stellar *OmnichainWalletStellar
}

type OmnichainWalletEvm struct {
	c   *Client
	Eoa *OmnichainWalletEvmEoa
}

type OmnichainWalletEvmEoa struct {
	c          *Client
	Operations *OmnichainWalletEvmEoaOperations
	Transfer   *OmnichainWalletEvmEoaTransfer
}

type OmnichainWalletEvmEoaOperations struct {
	c *Client
}

type OmnichainWalletEvmEoaTransfer struct {
	c *Client
}

type OmnichainWalletStellar struct {
	c   *Client
	Eoa *OmnichainWalletStellarEoa
}

type OmnichainWalletStellarEoa struct {
	c          *Client
	Operations *OmnichainWalletStellarEoaOperations
	Transfer   *OmnichainWalletStellarEoaTransfer
}

type OmnichainWalletStellarEoaOperations struct {
	c *Client
}

type OmnichainWalletStellarEoaTransfer struct {
	c *Client
}

type Simulation struct {
	c           *Client
	Card        *SimulationCard
	Collection  *SimulationCollection
	DirectDebit *SimulationDirectDebit
	Transfer    *SimulationTransfer
}

type SimulationCard struct {
	c *Client
}

type SimulationCollection struct {
	c *Client
}

type SimulationDirectDebit struct {
	c *Client
}

type SimulationTransfer struct {
	c *Client
}

type User struct {
	c            *Client
	Account      *UserAccount
	Misc         *UserMisc
	Subscription *UserSubscription
	Verification *UserVerification
}

type UserAccount struct {
	c          *Client
	Business   *UserAccountBusiness
	Individual *UserAccountIndividual
}

type UserAccountBusiness struct {
	c *Client
}

type UserAccountIndividual struct {
	c *Client
}

type UserMisc struct {
	c *Client
}

type UserSubscription struct {
	c *Client
}

type UserVerification struct {
	c *Client
}

type Verification struct {
	c *Client
}

type Wallet struct {
	c          *Client
	Channel    *WalletChannel
	Collection *WalletCollection
	Exchange   *WalletExchange
	Mpwt       *WalletMpwt
	Transfer   *WalletTransfer
}

type WalletChannel struct {
	c                 *Client
	TransferRecipient *WalletChannelTransferRecipient
	VirtualAccount    *WalletChannelVirtualAccount
	WalletAddress     *WalletChannelWalletAddress
}

type WalletChannelTransferRecipient struct {
	c             *Client
	Bank          *WalletChannelTransferRecipientBank
	MobileMoney   *WalletChannelTransferRecipientMobileMoney
	WalletAddress *WalletChannelTransferRecipientWalletAddress
}

type WalletChannelTransferRecipientBank struct {
	c *Client
}

type WalletChannelTransferRecipientMobileMoney struct {
	c *Client
}

type WalletChannelTransferRecipientWalletAddress struct {
	c *Client
}

type WalletChannelVirtualAccount struct {
	c *Client
}

type WalletChannelWalletAddress struct {
	c *Client
}

type WalletCollection struct {
	c      *Client
	Method *WalletCollectionMethod
}

type WalletCollectionMethod struct {
	c *Client
}

type WalletExchange struct {
	c              *Client
	Cryptotocrypto *WalletExchangeCryptotocrypto
	Cryptotofiat   *WalletExchangeCryptotofiat
	Fiattocrypto   *WalletExchangeFiattocrypto
	Fiattofiat     *WalletExchangeFiattofiat
}

type WalletExchangeCryptotocrypto struct {
	c        *Client
	External *WalletExchangeCryptotocryptoExternal
}

type WalletExchangeCryptotocryptoExternal struct {
	c      *Client
	Bridge *WalletExchangeCryptotocryptoExternalBridge
	Swap   *WalletExchangeCryptotocryptoExternalSwap
}

type WalletExchangeCryptotocryptoExternalBridge struct {
	c    *Client
	Usdc *WalletExchangeCryptotocryptoExternalBridgeUsdc
}

type WalletExchangeCryptotocryptoExternalBridgeUsdc struct {
	c *Client
}

type WalletExchangeCryptotocryptoExternalSwap struct {
	c *Client
}

type WalletExchangeCryptotofiat struct {
	c        *Client
	External *WalletExchangeCryptotofiatExternal
}

type WalletExchangeCryptotofiatExternal struct {
	c *Client
}

type WalletExchangeFiattocrypto struct {
	c      *Client
	Wallet *WalletExchangeFiattocryptoWallet
}

type WalletExchangeFiattocryptoWallet struct {
	c *Client
}

type WalletExchangeFiattofiat struct {
	c      *Client
	Wallet *WalletExchangeFiattofiatWallet
}

type WalletExchangeFiattofiatWallet struct {
	c *Client
}

type WalletMpwt struct {
	c *Client
}

type WalletTransfer struct {
	c      *Client
	Cross  *WalletTransferCross
	Direct *WalletTransferDirect
}

type WalletTransferCross struct {
	c *Client
}

type WalletTransferDirect struct {
	c *Client
}

func newAccount(c *Client) *Account {
	return &Account{
		c:        c,
		Merchant: newAccountMerchant(c),
	}
}

func newAccountMerchant(c *Client) *AccountMerchant {
	return &AccountMerchant{
		c:       c,
		Product: newAccountMerchantProduct(c),
		Team:    newAccountMerchantTeam(c),
	}
}

func newAccountMerchantProduct(c *Client) *AccountMerchantProduct {
	return &AccountMerchantProduct{
		c:            c,
		Subscription: newAccountMerchantProductSubscription(c),
	}
}

func newAccountMerchantProductSubscription(c *Client) *AccountMerchantProductSubscription {
	return &AccountMerchantProductSubscription{
		c: c,
	}
}

func newAccountMerchantTeam(c *Client) *AccountMerchantTeam {
	return &AccountMerchantTeam{
		c:          c,
		Permission: newAccountMerchantTeamPermission(c),
	}
}

func newAccountMerchantTeamPermission(c *Client) *AccountMerchantTeamPermission {
	return &AccountMerchantTeamPermission{
		c: c,
	}
}

func newAuthentication(c *Client) *Authentication {
	return &Authentication{
		c: c,
	}
}

func newCard(c *Client) *Card {
	return &Card{
		c: c,
	}
}

func newMisc(c *Client) *Misc {
	return &Misc{
		c: c,
	}
}

func newOmnichain(c *Client) *Omnichain {
	return &Omnichain{
		c:            c,
		Storage:      newOmnichainStorage(c),
		Tokenization: newOmnichainTokenization(c),
		Wallet:       newOmnichainWallet(c),
	}
}

func newOmnichainStorage(c *Client) *OmnichainStorage {
	return &OmnichainStorage{
		c: c,
	}
}

func newOmnichainTokenization(c *Client) *OmnichainTokenization {
	return &OmnichainTokenization{
		c:       c,
		Erc1155: newOmnichainTokenizationErc1155(c),
		Erc20:   newOmnichainTokenizationErc20(c),
		Erc721:  newOmnichainTokenizationErc721(c),
		Sep41:   newOmnichainTokenizationSep41(c),
	}
}

func newOmnichainTokenizationErc1155(c *Client) *OmnichainTokenizationErc1155 {
	return &OmnichainTokenizationErc1155{
		c: c,
	}
}

func newOmnichainTokenizationErc20(c *Client) *OmnichainTokenizationErc20 {
	return &OmnichainTokenizationErc20{
		c: c,
	}
}

func newOmnichainTokenizationErc721(c *Client) *OmnichainTokenizationErc721 {
	return &OmnichainTokenizationErc721{
		c: c,
	}
}

func newOmnichainTokenizationSep41(c *Client) *OmnichainTokenizationSep41 {
	return &OmnichainTokenizationSep41{
		c: c,
	}
}

func newOmnichainWallet(c *Client) *OmnichainWallet {
	return &OmnichainWallet{
		c:       c,
		Evm:     newOmnichainWalletEvm(c),
		Stellar: newOmnichainWalletStellar(c),
	}
}

func newOmnichainWalletEvm(c *Client) *OmnichainWalletEvm {
	return &OmnichainWalletEvm{
		c:   c,
		Eoa: newOmnichainWalletEvmEoa(c),
	}
}

func newOmnichainWalletEvmEoa(c *Client) *OmnichainWalletEvmEoa {
	return &OmnichainWalletEvmEoa{
		c:          c,
		Operations: newOmnichainWalletEvmEoaOperations(c),
		Transfer:   newOmnichainWalletEvmEoaTransfer(c),
	}
}

func newOmnichainWalletEvmEoaOperations(c *Client) *OmnichainWalletEvmEoaOperations {
	return &OmnichainWalletEvmEoaOperations{
		c: c,
	}
}

func newOmnichainWalletEvmEoaTransfer(c *Client) *OmnichainWalletEvmEoaTransfer {
	return &OmnichainWalletEvmEoaTransfer{
		c: c,
	}
}

func newOmnichainWalletStellar(c *Client) *OmnichainWalletStellar {
	return &OmnichainWalletStellar{
		c:   c,
		Eoa: newOmnichainWalletStellarEoa(c),
	}
}

func newOmnichainWalletStellarEoa(c *Client) *OmnichainWalletStellarEoa {
	return &OmnichainWalletStellarEoa{
		c:          c,
		Operations: newOmnichainWalletStellarEoaOperations(c),
		Transfer:   newOmnichainWalletStellarEoaTransfer(c),
	}
}

func newOmnichainWalletStellarEoaOperations(c *Client) *OmnichainWalletStellarEoaOperations {
	return &OmnichainWalletStellarEoaOperations{
		c: c,
	}
}

func newOmnichainWalletStellarEoaTransfer(c *Client) *OmnichainWalletStellarEoaTransfer {
	return &OmnichainWalletStellarEoaTransfer{
		c: c,
	}
}

func newSimulation(c *Client) *Simulation {
	return &Simulation{
		c:           c,
		Card:        newSimulationCard(c),
		Collection:  newSimulationCollection(c),
		DirectDebit: newSimulationDirectDebit(c),
		Transfer:    newSimulationTransfer(c),
	}
}

func newSimulationCard(c *Client) *SimulationCard {
	return &SimulationCard{
		c: c,
	}
}

func newSimulationCollection(c *Client) *SimulationCollection {
	return &SimulationCollection{
		c: c,
	}
}

func newSimulationDirectDebit(c *Client) *SimulationDirectDebit {
	return &SimulationDirectDebit{
		c: c,
	}
}

func newSimulationTransfer(c *Client) *SimulationTransfer {
	return &SimulationTransfer{
		c: c,
	}
}

func newUser(c *Client) *User {
	return &User{
		c:            c,
		Account:      newUserAccount(c),
		Misc:         newUserMisc(c),
		Subscription: newUserSubscription(c),
		Verification: newUserVerification(c),
	}
}

func newUserAccount(c *Client) *UserAccount {
	return &UserAccount{
		c:          c,
		Business:   newUserAccountBusiness(c),
		Individual: newUserAccountIndividual(c),
	}
}

func newUserAccountBusiness(c *Client) *UserAccountBusiness {
	return &UserAccountBusiness{
		c: c,
	}
}

func newUserAccountIndividual(c *Client) *UserAccountIndividual {
	return &UserAccountIndividual{
		c: c,
	}
}

func newUserMisc(c *Client) *UserMisc {
	return &UserMisc{
		c: c,
	}
}

func newUserSubscription(c *Client) *UserSubscription {
	return &UserSubscription{
		c: c,
	}
}

func newUserVerification(c *Client) *UserVerification {
	return &UserVerification{
		c: c,
	}
}

func newVerification(c *Client) *Verification {
	return &Verification{
		c: c,
	}
}

func newWallet(c *Client) *Wallet {
	return &Wallet{
		c:          c,
		Channel:    newWalletChannel(c),
		Collection: newWalletCollection(c),
		Exchange:   newWalletExchange(c),
		Mpwt:       newWalletMpwt(c),
		Transfer:   newWalletTransfer(c),
	}
}

func newWalletChannel(c *Client) *WalletChannel {
	return &WalletChannel{
		c:                 c,
		TransferRecipient: newWalletChannelTransferRecipient(c),
		VirtualAccount:    newWalletChannelVirtualAccount(c),
		WalletAddress:     newWalletChannelWalletAddress(c),
	}
}

func newWalletChannelTransferRecipient(c *Client) *WalletChannelTransferRecipient {
	return &WalletChannelTransferRecipient{
		c:             c,
		Bank:          newWalletChannelTransferRecipientBank(c),
		MobileMoney:   newWalletChannelTransferRecipientMobileMoney(c),
		WalletAddress: newWalletChannelTransferRecipientWalletAddress(c),
	}
}

func newWalletChannelTransferRecipientBank(c *Client) *WalletChannelTransferRecipientBank {
	return &WalletChannelTransferRecipientBank{
		c: c,
	}
}

func newWalletChannelTransferRecipientMobileMoney(c *Client) *WalletChannelTransferRecipientMobileMoney {
	return &WalletChannelTransferRecipientMobileMoney{
		c: c,
	}
}

func newWalletChannelTransferRecipientWalletAddress(c *Client) *WalletChannelTransferRecipientWalletAddress {
	return &WalletChannelTransferRecipientWalletAddress{
		c: c,
	}
}

func newWalletChannelVirtualAccount(c *Client) *WalletChannelVirtualAccount {
	return &WalletChannelVirtualAccount{
		c: c,
	}
}

func newWalletChannelWalletAddress(c *Client) *WalletChannelWalletAddress {
	return &WalletChannelWalletAddress{
		c: c,
	}
}

func newWalletCollection(c *Client) *WalletCollection {
	return &WalletCollection{
		c:      c,
		Method: newWalletCollectionMethod(c),
	}
}

func newWalletCollectionMethod(c *Client) *WalletCollectionMethod {
	return &WalletCollectionMethod{
		c: c,
	}
}

func newWalletExchange(c *Client) *WalletExchange {
	return &WalletExchange{
		c:              c,
		Cryptotocrypto: newWalletExchangeCryptotocrypto(c),
		Cryptotofiat:   newWalletExchangeCryptotofiat(c),
		Fiattocrypto:   newWalletExchangeFiattocrypto(c),
		Fiattofiat:     newWalletExchangeFiattofiat(c),
	}
}

func newWalletExchangeCryptotocrypto(c *Client) *WalletExchangeCryptotocrypto {
	return &WalletExchangeCryptotocrypto{
		c:        c,
		External: newWalletExchangeCryptotocryptoExternal(c),
	}
}

func newWalletExchangeCryptotocryptoExternal(c *Client) *WalletExchangeCryptotocryptoExternal {
	return &WalletExchangeCryptotocryptoExternal{
		c:      c,
		Bridge: newWalletExchangeCryptotocryptoExternalBridge(c),
		Swap:   newWalletExchangeCryptotocryptoExternalSwap(c),
	}
}

func newWalletExchangeCryptotocryptoExternalBridge(c *Client) *WalletExchangeCryptotocryptoExternalBridge {
	return &WalletExchangeCryptotocryptoExternalBridge{
		c:    c,
		Usdc: newWalletExchangeCryptotocryptoExternalBridgeUsdc(c),
	}
}

func newWalletExchangeCryptotocryptoExternalBridgeUsdc(c *Client) *WalletExchangeCryptotocryptoExternalBridgeUsdc {
	return &WalletExchangeCryptotocryptoExternalBridgeUsdc{
		c: c,
	}
}

func newWalletExchangeCryptotocryptoExternalSwap(c *Client) *WalletExchangeCryptotocryptoExternalSwap {
	return &WalletExchangeCryptotocryptoExternalSwap{
		c: c,
	}
}

func newWalletExchangeCryptotofiat(c *Client) *WalletExchangeCryptotofiat {
	return &WalletExchangeCryptotofiat{
		c:        c,
		External: newWalletExchangeCryptotofiatExternal(c),
	}
}

func newWalletExchangeCryptotofiatExternal(c *Client) *WalletExchangeCryptotofiatExternal {
	return &WalletExchangeCryptotofiatExternal{
		c: c,
	}
}

func newWalletExchangeFiattocrypto(c *Client) *WalletExchangeFiattocrypto {
	return &WalletExchangeFiattocrypto{
		c:      c,
		Wallet: newWalletExchangeFiattocryptoWallet(c),
	}
}

func newWalletExchangeFiattocryptoWallet(c *Client) *WalletExchangeFiattocryptoWallet {
	return &WalletExchangeFiattocryptoWallet{
		c: c,
	}
}

func newWalletExchangeFiattofiat(c *Client) *WalletExchangeFiattofiat {
	return &WalletExchangeFiattofiat{
		c:      c,
		Wallet: newWalletExchangeFiattofiatWallet(c),
	}
}

func newWalletExchangeFiattofiatWallet(c *Client) *WalletExchangeFiattofiatWallet {
	return &WalletExchangeFiattofiatWallet{
		c: c,
	}
}

func newWalletMpwt(c *Client) *WalletMpwt {
	return &WalletMpwt{
		c: c,
	}
}

func newWalletTransfer(c *Client) *WalletTransfer {
	return &WalletTransfer{
		c:      c,
		Cross:  newWalletTransferCross(c),
		Direct: newWalletTransferDirect(c),
	}
}

func newWalletTransferCross(c *Client) *WalletTransferCross {
	return &WalletTransferCross{
		c: c,
	}
}

func newWalletTransferDirect(c *Client) *WalletTransferDirect {
	return &WalletTransferDirect{
		c: c,
	}
}

func (s *AccountMerchant) ResendWebhook(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/resend-webhook", body, true)
}

func (s *AccountMerchant) UpdateNotification(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/update-notification", body, true)
}

func (s *AccountMerchantProduct) Update(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/product/update", body, true)
}

func (s *AccountMerchantProduct) View(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/product/view", body, true)
}

func (s *AccountMerchantProductSubscription) Class(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/product/subscription/class", body, true)
}

func (s *AccountMerchantProductSubscription) ClassProductType(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/product/subscription/class-product-type", body, true)
}

func (s *AccountMerchantProductSubscription) Finalise(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/product/subscription/finalise", body, true)
}

func (s *AccountMerchantProductSubscription) History(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/product/subscription/history", body, true)
}

func (s *AccountMerchantProductSubscription) Initiate(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/product/subscription/initiate", body, true)
}

func (s *AccountMerchantTeam) Add(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/team/add", body, true)
}

func (s *AccountMerchantTeam) Logs(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/team/logs", body, true)
}

func (s *AccountMerchantTeam) Remove(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/team/remove", body, true)
}

func (s *AccountMerchantTeam) Update(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/team/update", body, true)
}

func (s *AccountMerchantTeam) View(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/team/view", body, true)
}

func (s *AccountMerchantTeamPermission) Add(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/team/permission/add", body, true)
}

func (s *AccountMerchantTeamPermission) Remove(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/team/permission/remove", body, true)
}

func (s *AccountMerchantTeamPermission) Update(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/team/permission/update", body, true)
}

func (s *AccountMerchantTeamPermission) View(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "account/merchant/team/permission/view", body, true)
}

func (s *Authentication) Connect(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "authentication/connect", body, true)
}

func (s *Authentication) Initiate(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "authentication/initiate", body, false)
}

func (s *Authentication) SignOut(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "authentication/sign-out", body, true)
}

func (s *Authentication) Validate(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "authentication/validate", body, true)
}

func (s *Card) Close(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "card/close", body, true)
}

func (s *Card) Create(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "card/create", body, true)
}

func (s *Card) Details(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "card/details", body, true)
}

func (s *Card) Freeze(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "card/freeze", body, true)
}

func (s *Card) Fund(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "card/fund", body, true)
}

func (s *Card) Resubscribe(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "card/resubscribe", body, true)
}

func (s *Card) Supported(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "card/supported", body, true)
}

func (s *Card) Unfreeze(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "card/unfreeze", body, true)
}

func (s *Card) UpdateSecuritySettings(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "card/update-security-settings", body, true)
}

func (s *Card) Withdraw(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "card/withdraw", body, true)
}

func (s *Misc) Actions(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "actions", body, true)
}

func (s *Misc) Codes(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "codes", body, true)
}

func (s *Misc) Countries(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "countries", body, true)
}

func (s *Misc) FileUpload(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "file-upload", body, true)
}

func (s *Misc) PaymentPurpose(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "payment-purpose", body, true)
}

func (s *Misc) Products(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "products", body, true)
}

func (s *Misc) StateProvinceRegion(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "state-province-region", body, true)
}

func (s *Misc) SupportedCurrencies(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "supported-currencies", body, true)
}

func (s *OmnichainStorage) Details(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/storage/details", body, true)
}

func (s *OmnichainStorage) Upload(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/storage/upload", body, true)
}

func (s *OmnichainTokenizationErc1155) Balance(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc1155/balance", body, true)
}

func (s *OmnichainTokenizationErc1155) Burn(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc1155/burn", body, true)
}

func (s *OmnichainTokenizationErc1155) Deploy(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc1155/deploy", body, true)
}

func (s *OmnichainTokenizationErc1155) Mint(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc1155/mint", body, true)
}

func (s *OmnichainTokenizationErc1155) Transactions(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc1155/transactions", body, true)
}

func (s *OmnichainTokenizationErc1155) Transfer(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc1155/transfer", body, true)
}

func (s *OmnichainTokenizationErc20) Balance(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc20/balance", body, true)
}

func (s *OmnichainTokenizationErc20) Burn(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc20/burn", body, true)
}

func (s *OmnichainTokenizationErc20) Chains(body map[string]any) (map[string]any, error) {
	return s.c.Request("GET", "omnichain/tokenization/erc20/chains", body, true)
}

func (s *OmnichainTokenizationErc20) Deploy(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc20/deploy", body, true)
}

func (s *OmnichainTokenizationErc20) Mint(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc20/mint", body, true)
}

func (s *OmnichainTokenizationErc20) Transactions(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc20/transactions", body, true)
}

func (s *OmnichainTokenizationErc20) Transfer(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc20/transfer", body, true)
}

func (s *OmnichainTokenizationErc721) Balance(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc721/balance", body, true)
}

func (s *OmnichainTokenizationErc721) Burn(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc721/burn", body, true)
}

func (s *OmnichainTokenizationErc721) Deploy(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc721/deploy", body, true)
}

func (s *OmnichainTokenizationErc721) Mint(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc721/mint", body, true)
}

func (s *OmnichainTokenizationErc721) Transactions(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc721/transactions", body, true)
}

func (s *OmnichainTokenizationErc721) Transfer(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/erc721/transfer", body, true)
}

func (s *OmnichainTokenizationSep41) Balance(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/sep41/balance", body, true)
}

func (s *OmnichainTokenizationSep41) Burn(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/sep41/burn", body, true)
}

func (s *OmnichainTokenizationSep41) Deploy(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/sep41/deploy", body, true)
}

func (s *OmnichainTokenizationSep41) Mint(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/sep41/mint", body, true)
}

func (s *OmnichainTokenizationSep41) Transactions(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/sep41/transactions", body, true)
}

func (s *OmnichainTokenizationSep41) Transfer(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/tokenization/sep41/transfer", body, true)
}

func (s *OmnichainWalletEvmEoa) Balance(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/wallet/evm/eoa/balance", body, true)
}

func (s *OmnichainWalletEvmEoa) Chains(body map[string]any) (map[string]any, error) {
	return s.c.Request("GET", "omnichain/wallet/evm/eoa/chains", body, true)
}

func (s *OmnichainWalletEvmEoa) CreateWallet(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/wallet/evm/eoa/create-wallet", body, true)
}

func (s *OmnichainWalletEvmEoa) GenerateAddress(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/wallet/evm/eoa/generate-address", body, true)
}

func (s *OmnichainWalletEvmEoa) SupportedStablecoin(body map[string]any) (map[string]any, error) {
	return s.c.Request("GET", "omnichain/wallet/evm/eoa/supported-stablecoin", body, true)
}

func (s *OmnichainWalletEvmEoa) Transactions(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/wallet/evm/eoa/transactions", body, true)
}

func (s *OmnichainWalletEvmEoaOperations) BroadcastTransaction(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/wallet/evm/eoa/operations/broadcast-transaction", body, true)
}

func (s *OmnichainWalletEvmEoaOperations) EstimateGas(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/wallet/evm/eoa/operations/estimate-gas", body, true)
}

func (s *OmnichainWalletEvmEoaOperations) SignTransaction(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/wallet/evm/eoa/operations/sign-transaction", body, true)
}

func (s *OmnichainWalletEvmEoaTransfer) NativeToken(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/wallet/evm/eoa/transfer/native-token", body, true)
}

func (s *OmnichainWalletEvmEoaTransfer) Stablecoin(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/wallet/evm/eoa/transfer/stablecoin", body, true)
}

func (s *OmnichainWalletStellarEoa) Balance(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/wallet/stellar/eoa/balance", body, true)
}

func (s *OmnichainWalletStellarEoa) CreateWallet(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/wallet/stellar/eoa/create-wallet", body, true)
}

func (s *OmnichainWalletStellarEoa) GenerateAddress(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/wallet/stellar/eoa/generate-address", body, true)
}

func (s *OmnichainWalletStellarEoa) SupportedStablecoin(body map[string]any) (map[string]any, error) {
	return s.c.Request("GET", "omnichain/wallet/stellar/eoa/supported-stablecoin", body, true)
}

func (s *OmnichainWalletStellarEoa) Transactions(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/wallet/stellar/eoa/transactions", body, true)
}

func (s *OmnichainWalletStellarEoaOperations) CreateTrustline(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/wallet/stellar/eoa/operations/create-trustline", body, true)
}

func (s *OmnichainWalletStellarEoaOperations) EstimateGas(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/wallet/stellar/eoa/operations/estimate-gas", body, true)
}

func (s *OmnichainWalletStellarEoaTransfer) NativeToken(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/wallet/stellar/eoa/transfer/native-token", body, true)
}

func (s *OmnichainWalletStellarEoaTransfer) Stablecoin(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "omnichain/wallet/stellar/eoa/transfer/stablecoin", body, true)
}

func (s *SimulationCard) Spend(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "simulation/card/spend", body, true)
}

func (s *SimulationCollection) InitiatedOrder(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "simulation/collection/initiated-order", body, true)
}

func (s *SimulationCollection) StaticVirtualAccount(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "simulation/collection/static-virtual-account", body, true)
}

func (s *SimulationCollection) StaticWalletAddress(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "simulation/collection/static-wallet-address", body, true)
}

func (s *SimulationDirectDebit) Status(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "simulation/direct-debit/status", body, true)
}

func (s *SimulationTransfer) InitiatedOrder(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "simulation/transfer/initiated-order", body, true)
}

func (s *UserAccountBusiness) Create(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/account/business/create", body, true)
}

func (s *UserAccountBusiness) Details(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/account/business/details", body, true)
}

func (s *UserAccountBusiness) Update(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/account/business/update", body, true)
}

func (s *UserAccountIndividual) Create(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/account/individual/create", body, true)
}

func (s *UserAccountIndividual) Details(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/account/individual/details", body, true)
}

func (s *UserAccountIndividual) Update(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/account/individual/update", body, true)
}

func (s *UserMisc) BusinessIndustryList(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/misc/business-industry-list", body, true)
}

func (s *UserMisc) BusinessProductServiceType(body map[string]any) (map[string]any, error) {
	return s.c.Request("GET", "user/misc/business-product-service-type", body, false)
}

func (s *UserMisc) BusinessRegistrationType(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/misc/business-registration-type", body, true)
}

func (s *UserMisc) BusinessTradeType(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/misc/business-trade-type", body, true)
}

func (s *UserMisc) Countries(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/misc/countries", body, true)
}

func (s *UserMisc) JobTypes(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/misc/job-types", body, true)
}

func (s *UserMisc) SourceOfFundsType(body map[string]any) (map[string]any, error) {
	return s.c.Request("GET", "user/misc/source-of-funds-type", body, false)
}

func (s *UserSubscription) Class(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/subscription/class", body, true)
}

func (s *UserSubscription) Initiate(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/subscription/initiate", body, true)
}

func (s *UserSubscription) Products(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/subscription/products", body, true)
}

func (s *UserSubscription) Requery(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/subscription/requery", body, true)
}

func (s *UserSubscription) Status(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/subscription/status", body, true)
}

func (s *UserSubscription) Transactions(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/subscription/transactions", body, true)
}

func (s *UserVerification) BulkProcess(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/verification/bulk-process", body, true)
}

func (s *UserVerification) Requirement(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/verification/requirement", body, true)
}

func (s *UserVerification) SingleProcess(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "user/verification/single-process", body, true)
}

func (s *Verification) Lookup(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "verification/lookup", body, true)
}

func (s *Wallet) Create(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/create", body, true)
}

func (s *Wallet) Details(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/details", body, true)
}

func (s *Wallet) Transactions(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/transactions", body, true)
}

func (s *Wallet) Update(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/update", body, true)
}

func (s *WalletChannelTransferRecipientBank) Add(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/transfer-recipient/bank/add", body, true)
}

func (s *WalletChannelTransferRecipientBank) Countries(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/transfer-recipient/bank/countries", body, true)
}

func (s *WalletChannelTransferRecipientBank) Details(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/transfer-recipient/bank/details", body, true)
}

func (s *WalletChannelTransferRecipientBank) Institutions(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/transfer-recipient/bank/institutions", body, true)
}

func (s *WalletChannelTransferRecipientBank) Requirements(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/transfer-recipient/bank/requirements", body, true)
}

func (s *WalletChannelTransferRecipientBank) Resolve(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/transfer-recipient/bank/resolve", body, true)
}

func (s *WalletChannelTransferRecipientBank) SampleAccount(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/transfer-recipient/bank/sample-account", body, true)
}

func (s *WalletChannelTransferRecipientBank) SupportedCurrencies(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/transfer-recipient/bank/supported-currencies", body, true)
}

func (s *WalletChannelTransferRecipientBank) SupportedRouteType(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/transfer-recipient/bank/supported-route-type", body, true)
}

func (s *WalletChannelTransferRecipientMobileMoney) Add(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/transfer-recipient/mobile-money/add", body, true)
}

func (s *WalletChannelTransferRecipientMobileMoney) Details(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/transfer-recipient/mobile-money/details", body, true)
}

func (s *WalletChannelTransferRecipientMobileMoney) SampleAccount(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/transfer-recipient/mobile-money/sample-account", body, true)
}

func (s *WalletChannelTransferRecipientMobileMoney) SupportedCountries(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/transfer-recipient/mobile-money/supported-countries", body, true)
}

func (s *WalletChannelTransferRecipientMobileMoney) SupportedCurrencies(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/transfer-recipient/mobile-money/supported-currencies", body, true)
}

func (s *WalletChannelTransferRecipientMobileMoney) SupportedNetworks(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/transfer-recipient/mobile-money/supported-networks", body, true)
}

func (s *WalletChannelTransferRecipientWalletAddress) Add(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/transfer-recipient/wallet-address/add", body, true)
}

func (s *WalletChannelTransferRecipientWalletAddress) Details(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/transfer-recipient/wallet-address/details", body, true)
}

func (s *WalletChannelVirtualAccount) Details(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/virtual-account/details", body, true)
}

func (s *WalletChannelVirtualAccount) Generate(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/virtual-account/generate", body, true)
}

func (s *WalletChannelVirtualAccount) Resubscribe(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/virtual-account/resubscribe", body, true)
}

func (s *WalletChannelVirtualAccount) SupportedCurrencies(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/virtual-account/supported-currencies", body, true)
}

func (s *WalletChannelVirtualAccount) SupportedInstitutions(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/virtual-account/supported-institutions", body, true)
}

func (s *WalletChannelWalletAddress) Details(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/wallet-address/details", body, true)
}

func (s *WalletChannelWalletAddress) Generate(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/wallet-address/generate", body, true)
}

func (s *WalletChannelWalletAddress) SupportedTokens(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/channel/wallet-address/supported-tokens", body, true)
}

func (s *WalletCollection) Calculator(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/calculator", body, true)
}

func (s *WalletCollection) DesignOptions(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/design-options", body, true)
}

func (s *WalletCollection) Info(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/info", body, true)
}

func (s *WalletCollection) Initiate(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/initiate", body, true)
}

func (s *WalletCollection) SupportedMethods(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/supported-methods", body, true)
}

func (s *WalletCollectionMethod) Authorize3ds(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/method/authorize-3ds", body, true)
}

func (s *WalletCollectionMethod) AuthorizeOtp(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/method/authorize-otp", body, true)
}

func (s *WalletCollectionMethod) Charge(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/method/charge", body, true)
}

func (s *WalletCollectionMethod) Confirmation(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/method/confirmation", body, true)
}

func (s *WalletCollectionMethod) Create(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/method/create", body, true)
}

func (s *WalletCollectionMethod) Details(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/method/details", body, true)
}

func (s *WalletCollectionMethod) Link(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/method/link", body, true)
}

func (s *WalletCollectionMethod) ResendOtp(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/method/resend-otp", body, true)
}

func (s *WalletCollectionMethod) SupportedCountries(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/method/supported-countries", body, true)
}

func (s *WalletCollectionMethod) SupportedCurrencies(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/method/supported-currencies", body, true)
}

func (s *WalletCollectionMethod) SupportedInstitutions(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/method/supported-institutions", body, true)
}

func (s *WalletCollectionMethod) SupportedNetworks(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/method/supported-networks", body, true)
}

func (s *WalletCollectionMethod) Terminate(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/collection/method/terminate", body, true)
}

func (s *WalletExchangeCryptotocryptoExternalBridgeUsdc) Fees(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/exchange/cryptotocrypto/external/bridge/usdc/fees", body, true)
}

func (s *WalletExchangeCryptotocryptoExternalBridgeUsdc) Initiate(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/exchange/cryptotocrypto/external/bridge/usdc/initiate", body, true)
}

func (s *WalletExchangeCryptotocryptoExternalBridgeUsdc) Status(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/exchange/cryptotocrypto/external/bridge/usdc/status", body, true)
}

func (s *WalletExchangeCryptotocryptoExternalBridgeUsdc) SupportedChains(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/exchange/cryptotocrypto/external/bridge/usdc/supported-chains", body, true)
}

func (s *WalletExchangeCryptotocryptoExternalSwap) Fees(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/exchange/cryptotocrypto/external/swap/fees", body, true)
}

func (s *WalletExchangeCryptotocryptoExternalSwap) Initiate(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/exchange/cryptotocrypto/external/swap/initiate", body, true)
}

func (s *WalletExchangeCryptotocryptoExternalSwap) Status(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/exchange/cryptotocrypto/external/swap/status", body, true)
}

func (s *WalletExchangeCryptotocryptoExternalSwap) SupportedPairs(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/exchange/cryptotocrypto/external/swap/supported-pairs", body, true)
}

func (s *WalletExchangeCryptotofiatExternal) Fees(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/exchange/cryptotofiat/external/fees", body, true)
}

func (s *WalletExchangeCryptotofiatExternal) Initiate(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/exchange/cryptotofiat/external/initiate", body, true)
}

func (s *WalletExchangeCryptotofiatExternal) SupportedPairs(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/exchange/cryptotofiat/external/supported-pairs", body, true)
}

func (s *WalletExchangeFiattocryptoWallet) Fees(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/exchange/fiattocrypto/wallet/fees", body, true)
}

func (s *WalletExchangeFiattocryptoWallet) Initiate(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/exchange/fiattocrypto/wallet/initiate", body, true)
}

func (s *WalletExchangeFiattocryptoWallet) SupportedPairs(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/exchange/fiattocrypto/wallet/supported-pairs", body, true)
}

func (s *WalletExchangeFiattofiatWallet) Fees(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/exchange/fiattofiat/wallet/fees", body, true)
}

func (s *WalletExchangeFiattofiatWallet) Initiate(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/exchange/fiattofiat/wallet/initiate", body, true)
}

func (s *WalletExchangeFiattofiatWallet) SupportedPairs(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/exchange/fiattofiat/wallet/supported-pairs", body, true)
}

func (s *WalletMpwt) Balance(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/mpwt/balance", body, true)
}

func (s *WalletMpwt) Fees(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/mpwt/fees", body, true)
}

func (s *WalletMpwt) Fund(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/mpwt/fund", body, true)
}

func (s *WalletTransferCross) Fees(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/transfer/cross/fees", body, true)
}

func (s *WalletTransferCross) Single(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/transfer/cross/single", body, true)
}

func (s *WalletTransferCross) SupportedRoutes(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/transfer/cross/supported-routes", body, true)
}

func (s *WalletTransferDirect) Bulk(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/transfer/direct/bulk", body, true)
}

func (s *WalletTransferDirect) Fees(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/transfer/direct/fees", body, true)
}

func (s *WalletTransferDirect) Queue(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/transfer/direct/queue", body, true)
}

func (s *WalletTransferDirect) Single(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/transfer/direct/single", body, true)
}

func (s *WalletTransferDirect) SupportedCurrencies(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/transfer/direct/supported-currencies", body, true)
}

func (s *WalletTransferDirect) SupportedNetwork(body map[string]any) (map[string]any, error) {
	return s.c.Request("POST", "wallet/transfer/direct/supported-network", body, true)
}
