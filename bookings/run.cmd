go build -o bookings ./cmd/web/*.go
./bookings -dbname=bookings -dbuser=postgres -dbpwd=postgres -production=false -cache=false