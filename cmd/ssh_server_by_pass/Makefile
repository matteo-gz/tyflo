ssh_test_rsa:
	docker build -t ssh_pass_on_docker .
	docker-compose  up -d
ssh_in:
	docker exec -ti sshd_pass_my bash
ssh_client:
	ssh -p 10026 root@127.0.0.1 # password: redhat