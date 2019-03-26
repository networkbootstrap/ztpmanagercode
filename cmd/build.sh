rm ezjunosztp-alpha.tar.gz
rm ezjunosztp-linux-alpha
GOOS=linux go build -o ezjunosztp-linux-alpha
tar czvf ezjunosztp-alpha.tar.gz ezjunosztp-linux-alpha templates
