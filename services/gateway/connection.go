package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GatewayConfig holds all parameters needed to connect to the Fabric network.
type GatewayConfig struct {
	MSP_ID        string
	CryptoPath    string
	PeerEndpoint  string
	TLSCertPath   string
	ChannelName   string
	ChaincodeName string
	IdentityCert  string
	IdentityKey   string
	gateway       *client.Gateway // stored after Connect for ledger queries
}

// NewGatewayFromEnv creates a GatewayConfig from environment variables.
func NewGatewayFromEnv() *GatewayConfig {
	mspID := getEnv("MSP_ID", "Org1MSP")
	cryptoPath := getEnv("CRYPTO_PATH", "./crypto-config/peerOrganizations/org1.nanayam.com")

	// For complaint system, default identity paths point to ACB
	defaultCert := filepath.Join(cryptoPath, "users", fmt.Sprintf("Admin@%s", filepath.Base(cryptoPath)), "msp", "signcerts", fmt.Sprintf("Admin@%s-cert.pem", filepath.Base(cryptoPath)))
	defaultKey := filepath.Join(cryptoPath, "users", fmt.Sprintf("Admin@%s", filepath.Base(cryptoPath)), "msp", "keystore", "priv_sk")

	return &GatewayConfig{
		MSP_ID:        mspID,
		CryptoPath:    cryptoPath,
		PeerEndpoint:  getEnv("PEER_ENDPOINT", "localhost:7051"),
		TLSCertPath:   getEnv("TLS_CERT_PATH", "./crypto-config/peerOrganizations/org1.nanayam.com/peers/peer0.org1.nanayam.com/tls/ca.crt"),
		ChannelName:   getEnv("FABRIC_CHANNEL", "mychannel"),
		ChaincodeName: getEnv("FABRIC_CHAINCODE", "basic"),
		IdentityCert:  getEnv("IDENTITY_CERT", defaultCert),
		IdentityKey:   getEnv("IDENTITY_KEY", defaultKey),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Connect establishes a gRPC connection and Fabric Gateway client.
func (cfg *GatewayConfig) Connect() (*client.Gateway, *grpc.ClientConn, error) {
	// 1. Load TLS credentials
	certPEM, err := os.ReadFile(cfg.TLSCertPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read TLS cert: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to decode TLS cert PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse TLS cert: %w", err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(cert)
	transportCreds := credentials.NewClientTLSFromCert(certPool, "")

	// 2. Create gRPC connection
	conn, err := grpc.Dial(cfg.PeerEndpoint, grpc.WithTransportCredentials(transportCreds))
	if err != nil {
		return nil, nil, fmt.Errorf("grpc dial: %w", err)
	}

	// 3. Load identity (certificate)
	certPEM, err = os.ReadFile(cfg.IdentityCert)
	if err != nil {
		return nil, nil, fmt.Errorf("read identity cert: %w", err)
	}

	block, _ = pem.Decode(certPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to decode identity cert PEM")
	}

	identityCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse identity cert: %w", err)
	}

	id, err := identity.NewX509Identity(cfg.MSP_ID, identityCert)
	if err != nil {
		return nil, nil, fmt.Errorf("new identity: %w", err)
	}

	// 4. Load private key for signing
	keyPEM, err := os.ReadFile(cfg.IdentityKey)
	if err != nil {
		return nil, nil, fmt.Errorf("read private key: %w", err)
	}

	sign, err := newSign(keyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("new signer: %w", err)
	}

	// 5. Create gateway client
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(conn),
		client.WithEvaluateTimeout(30*time.Second),
		client.WithEndorseTimeout(30*time.Second),
		client.WithSubmitTimeout(30*time.Second),
		client.WithCommitStatusTimeout(60*time.Second),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("connect gateway: %w", err)
	}

	cfg.gateway = gw
	return gw, conn, nil
}

// newSign creates a digital signature from a private key.
func newSign(keyPEM []byte) (identity.Sign, error) {
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key PEM")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err = x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
	}

	sign, err := identity.NewPrivateKeySign(privateKey.(*ecdsa.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("new private key sign: %w", err)
	}

	return sign, nil
}
