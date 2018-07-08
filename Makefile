all: secp256k1 ucoin

ucoin:
	go install github.com/tokenme/ucoin

secp256k1:
	cp -r dependencies/secp256k1/src vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/libsecp256k1/src;
	cp -r dependencies/secp256k1/include vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/libsecp256k1/include;

install:
	#rm -rf /opt/ucoin-static/*;
	#cp -r ui/build/dist/* /opt/ucoin-static/;
	cp -f /opt/go/bin/ucoin /usr/local/bin/;
	chmod a+x /usr/local/bin/ucoin;
