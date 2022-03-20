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

### --data-path
Will define where stored data is put (i.e. IP addresses for subdomains). All records will be saved here every second if they have been changed and when the application gets shut down.  
Type: `string`  
Default: `.`

## Routes
### `/`
If in debug mode, will output all registered records, otherwise prints `"Working"`.

### `/update/{domain}/{type}`
Updates a given record.  
#### URL parts
`domain` (`string`): defines the subdomain that is to be modified  
`type` (`"v4"` | `"v6"`): specifies whether an IPv4 or IPv6 record is to be changed.  
#### query parameters
`ipv4` (`string`) (only if `type` is `"v4"`): specifies the IPv4 address to be applied.  
`ipv6` (`string`) (only if `type` is `"v6"`): specifies the IPv6 address to be applied.  
`user` (`string`): username as specified in _auth file_.
`password` (`string`): password as specified in _auth file_.

#### examples
/update/**example.example**/**v4**?ipv4=**127.0.0.1**&user=**username**&password=**password**  

/update/**example.example**/**v6**?ipv6=**::1**&user=**username**&password=**password**  
