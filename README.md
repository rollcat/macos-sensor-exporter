# macos-sensor-exporter

This is a simple [Prometheus](https://prometheus.io) exporter for
macOS-specific metrics, obtained via the SMC (System Management
Controller); it can report values such as power consumption, battery
levels, temperature, fans, etc.

It exposes the metrics in <http://localhost:9101/metrics>.

You can build it, and set it up to run from launchd by doing the following:

1. `go get && go build -o ~/bin/macos-sensor-exporter`
2. Set e.g. `DNS=test.example` (you can use a domain you own; `*.test`
   is a reserved TLD can be used internally by anyone);
3. Run the following snippet:

```sh
cat > ~/Library/LaunchAgents/$DNS.macos-sensor-exporter.plist <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key>
    <string>$DNS.macos-sensor-exporter.plist</string>
    <key>ProgramArguments</key>
    <array>
      <string>$HOME/bin/macos-sensor-exporter</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
  </dict>
</plist>
EOF
```

4. Load the service with:
   `launchctl load ~/Library/LaunchAgents/$DNS.macos-sensor-exporter.plist`

You can configure Prometheus to scrape this exporter like so:

```yaml
- job_name: "macos-sensor-exporter"
  scrape_timeout: 3s
  static_configs:
  - targets: ["localhost:9101"]
```

On my Mac Mini M1, the average scrape time is around 2s, so I've set a
larger timeout value (3s) than for the rest of my setup (1s).
