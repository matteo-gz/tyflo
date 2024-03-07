ssh_test_rsa:
	docker build -t ssh_on_docker .
	docker-compose  up -d
ssh_in:
	docker exec -ti sshd_my bash
ssh_client:
	ssh -i key/id_rsa 127.0.0.1:10025
ssh_fix:
	chmod 600 key/id_rsa