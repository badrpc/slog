# slog

GoLang package slog provides alternative syslog client API. An internal
syslog writer used to send messages to a syslog service with options
to tune it.

slog is not an officially supported Google product.

## Example

``` go
	// ...

        if f, err := slog.ParseFacility(syslogFacility); err != nil {
                slog.Err(err)
                os.Exit(exTempFail)
        } else {
                slog.Init(slog.WithFacility(f))
        }

	// ...

        slog.Info("Job ID: ", jobId)
        slog.Info("Message-Id: ", mime.GetHeader("Message-Id"))

	// ...
```
