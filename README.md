# gertdns
A DynDNS server meant for gertroot

Running:
```sh
go run main.go
```

Bullding:
```sh
go build
```

## Config
Default `conf.toml`
```toml
[DNS]
Port    = 5353               # DNS server port
Host    = '0.0.0.0'          # DNS server host
Domains = ['example.com.']   # enabled domains, suffix with a .

[HTTP]
Port           = 8080        # HTTP server port
Host           = '127.0.0.1' # HTTP server host
Socket         = ''          # HTTP unix socket
SocketFileMode = 420         # File mode for HTTP unix socket in decimal (420 = 0644)
```

## Users
Default `auth.toml`
```toml
[someusername]  # user name of the user
Password = '1234'                     # password of the user
Hashed   = false                      # false if omitted; if false, password will be hashed
Domains  = ["subdomain.example.com."] # domains the user can register, suffix with a .

# ..
```

## Flags
### --enable-debug-mode
Will output all registered records on the index page of the HTTP server.  
Type: `bool`  
Default: `false`

### --config-file
Will define what config file should be used.  
Type: `string`  
Default: `conf.toml`


### --auth-file
Will define what file should be used to define users that can log in.  
Type: `string`  
Default: `auth.toml`
