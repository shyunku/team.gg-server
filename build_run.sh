go build ./main.go
tail -f output.log
sudo supervisorctl restart team.gg-server