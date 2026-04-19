# vaultwatch

A CLI tool to audit and alert on expiring secrets in HashiCorp Vault.

## Installation

```bash
go install github.com/yourusername/vaultwatch@latest
```

Or download a pre-built binary from the [releases page](https://github.com/yourusername/vaultwatch/releases).

## Usage

Set your Vault address and token, then run an audit:

```bash
export VAULT_ADDR="https://vault.example.com"
export VAULT_TOKEN="s.yourtoken"

# Audit all secrets and alert on those expiring within 30 days
vaultwatch audit --path secret/ --warn-within 30d

# Output results as JSON
vaultwatch audit --path secret/ --format json

# Watch continuously and send alerts
vaultwatch watch --path secret/ --interval 1h --alert slack
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--path` | Vault path to audit | `secret/` |
| `--warn-within` | Alert threshold duration | `30d` |
| `--format` | Output format (`text`, `json`) | `text` |
| `--interval` | Watch mode poll interval | `1h` |
| `--alert` | Alert backend (`slack`, `pagerduty`) | none |

## Configuration

VaultWatch can be configured via a `vaultwatch.yaml` file:

```yaml
vault_addr: https://vault.example.com
warn_within: 30d
alert:
  backend: slack
  webhook_url: https://hooks.slack.com/services/...
```

## Requirements

- Go 1.21+
- HashiCorp Vault 1.10+

## License

MIT — see [LICENSE](LICENSE) for details.