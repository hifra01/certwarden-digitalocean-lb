package main

import (
	"context"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/digitalocean/godo"
)

func main() {
	// 1. Get Environment Variables
	doToken := os.Getenv("DO_TOKEN")
	lbID := os.Getenv("LB_ID")
	// lbName := os.Getenv("LB_NAME")
	// lbRegion := os.Getenv("LB_REGION")

	cwCertPEM := os.Getenv("CW_CERTIFICATE_PEM")
	cwPrivKeyPEM := os.Getenv("CW_PRIVATE_KEY_PEM")
	cwCertCommonName := os.Getenv("CW_CERTIFICATE_COMMON_NAME")

	if doToken == "" || lbID == "" || cwCertPEM == "" || cwPrivKeyPEM == "" || cwCertCommonName == "" {
		log.Fatal("Missing required environment variables. Ensure DO_TOKEN, LB_ID, CW_CERTIFICATE_PEM, CW_PRIVATE_KEY_PEM, and CW_CERTIFICATE_COMMON_NAME are set.")
	}

	// Just checking them as requested by the user, although AsRequest() will preserve the actual name/region.
	// if lbName != "" {
	// 	log.Printf("LB Name from env: %s", lbName)
	// }
	// if lbRegion != "" {
	// 	log.Printf("LB Region from env: %s", lbRegion)
	// }

	// 2. Get leaf certificate (first block only to fulfill DigitalOcean requirement)
	block, _ := pem.Decode([]byte(cwCertPEM))
	if block == nil {
		log.Fatal("Failed to parse PEM block from CW_CERTIFICATE_PEM")
	}
	leafCertPEM := string(pem.EncodeToMemory(block))

	// 3. Get the private key (already in correct format)
	privKeyPEM := cwPrivKeyPEM

	// 4. Create the certificate on DigitalOcean
	certName := fmt.Sprintf("%s-%s", cwCertCommonName, time.Now().Format("200601021504"))
	log.Printf("Starting DigitalOcean Load Balancer update for %s...", cwCertCommonName)

	client := godo.NewFromToken(doToken)
	ctx := context.Background()

	certReq := &godo.CertificateRequest{
		Name:            certName,
		Type:            "custom",
		PrivateKey:      privKeyPEM,
		LeafCertificate: leafCertPEM,
	}

	cert, _, err := client.Certificates.Create(ctx, certReq)
	if err != nil {
		log.Fatalf("ERROR: Failed to upload certificate to DigitalOcean API. Error: %v", err)
	}
	log.Printf("Successfully uploaded certificate to DigitalOcean. ID: %s", cert.ID)

	// 5. Update Load Balancer to use new certificate, keeping other info persisted
	lb, _, err := client.LoadBalancers.Get(ctx, lbID)
	if err != nil {
		log.Fatalf("ERROR: Failed to retrieve Load Balancer %s. Error: %v", lbID, err)
	}

	lbReq := lb.AsRequest()

	var newRules []godo.ForwardingRule
	foundHTTPS443 := false

	for _, rule := range lbReq.ForwardingRules {
		if rule.EntryProtocol == "https" && rule.EntryPort == 443 {
			// Replace this rule's certificate ID but keep other target configs
			rule.CertificateID = cert.ID
			foundHTTPS443 = true
		}
		newRules = append(newRules, rule)
	}

	if !foundHTTPS443 {
		// Add new rule for https 443 as fallback
		newRules = append(newRules, godo.ForwardingRule{
			EntryProtocol:  "https",
			EntryPort:      443,
			TargetProtocol: "http",
			TargetPort:     80,
			CertificateID:  cert.ID,
		})
	}

	lbReq.ForwardingRules = newRules

	_, _, err = client.LoadBalancers.Update(ctx, lbID, lbReq)
	if err != nil {
		log.Fatalf("ERROR: Failed to update Load Balancer. Error: %v", err)
	}

	log.Println("Load balancer updated successfully!")
}
