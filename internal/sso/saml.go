package sso

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"limiter.io/internal/models"
	"limiter.io/internal/repository"
)

type SAMLConfig struct {
	IDPEntityID    string `json:"idp_entity_id"`
	IDPSSOURL      string `json:"idp_sso_url"`
	IDPSSOBinding  string `json:"idp_sso_binding"`
	IDPPublicCert  string `json:"idp_public_cert"`
	SPEntityID     string `json:"sp_entity_id"`
	SPACSURL       string `json:"sp_acs_url"`
	SPPrivateKey   string `json:"sp_private_key"`
	SPPublicCert   string `json:"sp_public_cert"`
	OrganizationID string `json:"organization_id"`
	Enabled        bool   `json:"enabled"`
}

type OIDCConfig struct {
	IssuerURL      string `json:"issuer_url"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	RedirectURL    string `json:"redirect_url"`
	OrganizationID string `json:"organization_id"`
	Enabled        bool   `json:"enabled"`
	Scopes         string `json:"scopes"`
}

type SSOService struct {
	repo repository.SSORepository
}

func NewSSOService(repo repository.SSORepository) *SSOService {
	return &SSOService{repo: repo}
}

func (s *SSOService) SetSAMLConfig(cfg *SAMLConfig) error {
	m := &models.SAMLConfig{
		OrganizationID: cfg.OrganizationID,
		IDPEntityID:    cfg.IDPEntityID,
		IDPSSOURL:      cfg.IDPSSOURL,
		IDPSSOBinding:  cfg.IDPSSOBinding,
		IDPPublicCert:  cfg.IDPPublicCert,
		SPEntityID:     cfg.SPEntityID,
		SPACSURL:       cfg.SPACSURL,
		SPPrivateKey:   cfg.SPPrivateKey,
		SPPublicCert:   cfg.SPPublicCert,
		Enabled:        cfg.Enabled,
	}
	return s.repo.SaveSAMLConfig(context.Background(), m)
}

func (s *SSOService) GetSAMLConfig(orgID string) (*SAMLConfig, error) {
	m, err := s.repo.GetSAMLConfig(context.Background(), orgID)
	if err != nil {
		return nil, errors.New("SAML config not found")
	}
	return &SAMLConfig{
		OrganizationID: m.OrganizationID,
		IDPEntityID:    m.IDPEntityID,
		IDPSSOURL:      m.IDPSSOURL,
		IDPSSOBinding:  m.IDPSSOBinding,
		IDPPublicCert:  m.IDPPublicCert,
		SPEntityID:     m.SPEntityID,
		SPACSURL:       m.SPACSURL,
		SPPrivateKey:   m.SPPrivateKey,
		SPPublicCert:   m.SPPublicCert,
		Enabled:        m.Enabled,
	}, nil
}

func (s *SSOService) SetOIDCConfig(cfg *OIDCConfig) error {
	m := &models.OIDCConfig{
		OrganizationID: cfg.OrganizationID,
		IssuerURL:      cfg.IssuerURL,
		ClientID:       cfg.ClientID,
		ClientSecret:   cfg.ClientSecret,
		RedirectURL:    cfg.RedirectURL,
		Scopes:         cfg.Scopes,
		Enabled:        cfg.Enabled,
	}
	return s.repo.SaveOIDCConfig(context.Background(), m)
}

func (s *SSOService) GetOIDCConfig(orgID string) (*OIDCConfig, error) {
	m, err := s.repo.GetOIDCConfig(context.Background(), orgID)
	if err != nil {
		return nil, errors.New("OIDC config not found")
	}
	return &OIDCConfig{
		OrganizationID: m.OrganizationID,
		IssuerURL:      m.IssuerURL,
		ClientID:       m.ClientID,
		ClientSecret:   m.ClientSecret,
		RedirectURL:    m.RedirectURL,
		Scopes:         m.Scopes,
		Enabled:        m.Enabled,
	}, nil
}

func ParsePrivateKey(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func ParseCertificate(pemData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}
	return x509.ParseCertificate(block.Bytes)
}
