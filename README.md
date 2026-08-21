# Certwarden DigitalOcean Load Balancer Updater

A Go-based post-processing utility for [Certwarden](https://github.com/certwarden/certwarden) to automatically upload newly issued Let's Encrypt certificates to DigitalOcean and update a DigitalOcean Load Balancer to use them. 
## Features

- Safely extracts the leaf certificate block from Certwarden's standard PEM variables, fulfilling DigitalOcean's API constraints.
- Uploads the certificate directly to DigitalOcean with a unique name: `<Common_Name>-<Timestamp>`.
- Preserves all existing configurations on your DigitalOcean Load Balancer.
- Smoothly updates only the HTTPS (port 443) forwarding rule to use the new certificate. Leaves HTTP (port 80) and other rules fully intact.

## Configuration

This binary reads its configuration directly from environment variables. 

### DigitalOcean Variables
You must provide the following variables manually (e.g., in your Certwarden job configuration):

| Variable | Description |
| --- | --- |
| `DO_TOKEN` | Your DigitalOcean Personal Access Token with read/write access. |
| `LB_ID` | The ID of the DigitalOcean Load Balancer you wish to update. |


### Certwarden Variables
Certwarden allows you to inject dynamic certificate values into environment variables by using placeholders in the **Post Processing** configuration of your certificate.

You should configure the following environment variables for this binary to use:

| Key | Value | Description |
| --- | --- | --- |
| `CW_CERTIFICATE_PEM` | `{{CERTIFICATE_PEM}}` | The full certificate chain in PEM format. |
| `CW_PRIVATE_KEY_PEM` | `{{PRIVATE_KEY_PEM}}` | The private key for the certificate in PEM format. |
| `CW_CERTIFICATE_COMMON_NAME` | `{{CERTIFICATE_COMMON_NAME}}` | The primary domain/common name of the issued certificate. |

## Usage with Certwarden

1. Download the latest binary for your operating system from the [Releases page](../../releases).
2. Place the binary on your server where Certwarden runs. Make sure it is executable (`chmod +x certwarden-do-lb-linux-amd64`).
3. In Certwarden, under your certificate's **Post Processing** options, specify the absolute path to the binary (e.g., `/path/to/certwarden-do-lb-linux-amd64`). Certwarden will automatically detect that it is a binary.
4. Add the following Environment Variables in the Certwarden UI for this Post Processing step:
   - `DO_TOKEN` = `your_digital_ocean_api_token`
   - `LB_ID` = `your_load_balancer_id`
   - `CW_CERTIFICATE_PEM` = `{{CERTIFICATE_PEM}}`
   - `CW_PRIVATE_KEY_PEM` = `{{PRIVATE_KEY_PEM}}`
   - `CW_CERTIFICATE_COMMON_NAME` = `{{CERTIFICATE_COMMON_NAME}}`

## Building from Source

If you wish to compile the program yourself, ensure you have Go 1.21+ installed.

```bash
# Clone the repository
git clone https://github.com/your-username/certwarden-do-lb.git
cd certwarden-do-lb

# Build the binary
go build -o certwarden-do-lb main.go
```
