# Flight

This project aggregates multiple APIs and present a list of available flights sorted by price

## Configuration
This project uses `environment variables` to hold default values and secrets.
These values can be replaced directly in the app.env file or by passing them to `docker run -e`.

**I'm including the keys I used for development**, you can use those or create new ones. 
Amadeus and Sky should work fine, however Flight API uses development credits.
I've only have 16 left and every request consumes 2 credits.

Alternatively you can use `Secrets` or `Vault` to manage these settings, the server won't fail as long as the values are loaded.

### Certificates
I'm including of self signed certificates I used for development. Of course these wont work on prod. To use real certificates just drop them in
the `/certs` folder.

Ideally I would use TLS termination with an ingress controller.

## APIs
I manage to implement three APIs
- Flight API
- Amadeus Self-Service Flight
- SkyScanner through rapid API.

### API provider
All three APIS implement the Provider interface

```go
type Provider interface {
	Search(ctx context.Context, params SearchParams) ([]Info, error)
}
```

On the handler I create a gorutine for each, and call the Search method

```go
var wg sync.WaitGroup

wg.Add(len(providers))

for i, p := range providers {
    go func() {
        defer wg.Done()

        t := time.Now()

        log.Printf("sending request for provider # %v", i)

        info, err := p.Search(ctx, params)

        log.Printf("request for provider # %v took %s", i, time.Since(t))

        if err != nil {
            errCh <- err
            return
        }

        infoCh <- info
    }()
}

wg.Wait()
done <- true
```

The synchronization is handled in a select statement

```go
go func() {
    for {
        select {
        case err := <-errCh:
            log.Println(err)
            workingProviders--
        case info := <-infoCh:
            data = append(data, info...)
        case <-done:
            return
        }
    }
}()
```

## JWT Token
Very simple, I created a package to create and verify tokens using a secret key. Request are protected using a middleware

```go
var handler http.Handler = mux

mux := http.NewServeMux()
addRoutes(mux, providers, tokenMaker, config)

var handler http.Handler = mux

handler = jwtMiddleware(handler, tokenMaker)

return handler
```
## Tests
Only the token and http package have unit test, http been integration tests. 
The regression test covers every scenario, and verify the results based on individual APIs failing
making it possible to detect the outcome.

- The regression tests load data from mock files.
- The mock services use channels to simulate the behavior of failing.

```
go test ./... -cover -race
```

### UI
The front end is a build vue app served as static files by the http server.

