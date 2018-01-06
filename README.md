# go-altgains

## Installation

Requires [Go](https://golang.org/dl/) and [go-binance](https://github.com/pdepip/go-binance):

```bash
$ go get github.com/pdepip/go-binance/binance github.com/7AC/go-altgains
```

## Usage

```bash
$ go run $GOPATH/src/github.com/7AC/go-altgains/altgains.go
Coin  Total Balance  Available Balance  In Order  BTC Value  BTC Gain   USDT Value  USDT Gain  % Gain
ADA   305.694000     305.694000         0.000000  0.018449   +0.018274  313.79      +310.83    +104.82
BTC   0.155645       0.155645           0.000000  0.155645   -0.102273  412.05      -1739.56   -0.81
ENG   75.924000      75.924000          0.000000  0.024293   +0.021613  413.21      +367.61    +8.06
ETH   2.244211       2.244211           0.000000  0.131820   +0.043839  2242.13     +745.65    +0.50
LTC   14.983652      9.983652           5.000000  0.244234   -0.045397  4154.17     -772.16    -0.16               
NEO   0.999000       0.999000           0.000000  0.005827   +0.000349  99.11       +5.93      +0.06
OST   99.900000      99.900000          0.000000  0.004436   +0.004335  75.44       +73.73     +42.95
TRX   991.008000     991.008000         0.000000  0.009296   +0.009290  158.11      +158.01    +1621.93          
XRP   242.757000     242.757000         0.000000  0.037443   +0.036835  636.87      +626.52    +60.59            
                                                                              
Total BTC Value: 0.631443                                                                       
Total USDT Value: 10740.21                            
```

## Supported exchanges
For now only [Binance](http://binance.com).
