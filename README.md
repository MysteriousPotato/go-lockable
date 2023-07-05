<!-- PROJECT LOGO -->
<br />
<div style="align-content: center;">
  <!--<a href="https://github.com/MysteriousPotato/go-lockable">
    <img src="images/logo.png" alt="Logo" width="80" height="80">
  </a>-->

<h3 align="center">go-lockable</h3>

  <p align="center">
    key based lock
    <br />
</div>



<!-- ABOUT THE PROJECT -->
## About The Project

go-lockable provides a simple implementation for acquiring locks by key.

This can be useful when multiple goroutines need to manipulate a map atomically using async calls without blocking access to all keys or when access to certain ressource needs to be partially locked.

This pkg does not use any 3rd party dependencies, only std packages. 

<!-- GETTING STARTED -->
## Getting Started

### Installation

```sh
go get github.com/MysteriousPotato/go-lockable
```

<!-- USAGE EXAMPLES -->
## Usage
```go
package main


import (
	"github.com/MysteriousPotato/go-lockable"
)

func main() {
	// Using lockable like a mutex:
    lock := lockable.New[string]()
	lock.LockKey("potato")
    defer lock.UnlockKey("potato")
    
    // Using go-lockable built-in Map type:
    lockableMap := lockable.NewMap[string, int]()
    lockableMap.LockKey("potato")
    defer lockableMap.UnlockKey("potato")
    
    // Do async stuff....
    
    lockableMap.Store("potato", 10)
}

```

For more detailed examples, please refer to the [Documentation](https://pkg.go.dev/github.com/MysteriousPotato/go-lockable)_

<!-- ROADMAP -->
## Roadmap

See the [open issues](https://github.com/MysteriousPotato/go-lockable/issues) for a full list of proposed features (and known issues).

<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".

<!-- LICENSE -->
## License

Distributed under the MIT License. See [LICENSE](https://github.com/MysteriousPotato/go-lockable/blob/master/LICENSE) for more information.
