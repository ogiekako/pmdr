# Set up
Download your client_secret.json following [this instruction](https://developers.google.com/google-apps/calendar/quickstart/go).

# Add/Remove schedule
```go
go run add.go
```

adds 4 pomodoros.

```go
go run add.go --from 15:00 8
```

adds 8 pomodoros from 15:00.


```go
go run add.go --remove --from 14:00
```

removes the pomodoros from >= 14:00.
