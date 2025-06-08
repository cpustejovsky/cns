# Cloud Native Service


## Next Steps
Replace gorilla/mux with std lib

## Notes from the Cloud Native Go

* Idempotent (PUT requests and `x = 1` where it doesn't change no matter how many times is called)
    * Why use them?
        * safer
        * simpler
        * more declarative
    * Essential for a cloud native architecture
    * Mathematical Definition: func `f` is idempotent IFF `f(f(x)) = f(x)` for all `x`
* 
