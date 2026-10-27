# Consul's HTTP API for explicit registration.
# Run this after product-service, order-service, and api-gateway are running.
$services = @(
  @{ID="product-service-1"; Name="product-service"; Address="127.0.0.1"; Port=8081},
  @{ID="order-service-1"; Name="order-service"; Address="127.0.0.1"; Port=8080},
  @{ID="api-gateway-1"; Name="api-gateway"; Address="127.0.0.1"; Port=8000}
)
foreach ($s in $services) {
  $body = $s | ConvertTo-Json
  Invoke-RestMethod -Method Put -Uri "http://localhost:8500/v1/agent/service/register" -ContentType "application/json" -Body $body
}
Write-Host "Services registered. Check http://localhost:8500"