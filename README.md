# PHP Process Manager

This program starts multiple PHP development servers and load balances incoming requests between them using a round-robin algorithm. The PHP development servers listen on consecutive ports starting from a specified port number.

## Requirements

- Go 1.15 or later
- PHP 7.4 or later

## Installation

1. Clone the repository:
```
git clone https://github.com/myusername/php-process-manager
```

2. Build the program:

## Usage

```
./php-process-manager -process-port=8000 -balancer-port=9090
```


The `process-port` flag specifies the starting port number for the PHP development servers. The `balancer-port` flag specifies the port number for the load balancer.

Alternatively, you can set the `PROCESS_PORT` and `BALANCER_PORT` environment variables before running the program. If the flags are not specified, the environment variables will be used. If neither the flags nor the environment variables are specified, the default values of 8000 and 9090 will be used, respectively.

## License

This program is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
