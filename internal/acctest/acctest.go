// Copyright (c) David Roman
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"os"
	"testing"

	"github.com/davidroman0O/terraform-provider-turingpi/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ProtoV6ProviderFactories returns provider factories for acceptance testing.
var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"turingpi": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// PreCheck validates required environment variables for acceptance tests.
func PreCheck(t *testing.T) {
	if v := os.Getenv("TURINGPI_HOST"); v == "" {
		t.Fatal("TURINGPI_HOST must be set for acceptance tests")
	}
	if v := os.Getenv("TURINGPI_USERNAME"); v == "" {
		t.Fatal("TURINGPI_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("TURINGPI_PASSWORD"); v == "" {
		t.Fatal("TURINGPI_PASSWORD must be set for acceptance tests")
	}
}
