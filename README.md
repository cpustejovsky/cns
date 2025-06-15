# Cloud Native Service


## Next Steps
### File transaction logger
    * ~~Add tests~~
    * ~~Add Close method to gracefully close file~~
    * ~~Ensure write buffer is drained on close~~
    * ~~Encode keys and values; ensure white-space characters will parse correctly~~
    * Bound keys and values
    * Converted keys and values to a binary format
    * Solution to retaining deleted values indefinitely
### PostgreSQL transaction logger
    * create database if it doesn't exist
    * Close method to clean up open connections
    * drain events channel on close
    * indefinite growth of values
### TLS
    * Add TODOs
### Containerization
    * Publishable container that uses postgres
    * Documentation for Dockerfile
### Misc.
* ~~Replace gorilla/mux with std lib~~
## Notes from the Cloud Native Go

### Idempotent 
* PUT requests and `x = 1` where it doesn't change no matter how many times is called
* Why use them?
    * safer
    * simpler
    * more declarative
* Essential for a cloud native architecture
* Mathematical Definition: func `f` is idempotent IFF `f(f(x)) = f(x)` for all `x`
### stateless
* application state is worse than resource state
#### application state
* server-side data about the app or how the client is using the app
* example: session tracking
#### resource state
* state of a resource within a service
