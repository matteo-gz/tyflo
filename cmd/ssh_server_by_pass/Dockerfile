# https://kinsta.com/blog/ssh-into-docker-container/
FROM ubuntu:latest
# https://blog.csdn.net/qq_43753724/article/details/122463902
RUN sed -i s@/archive.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list && \
    sed -i s@/security.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list && \
    apt-get clean && \
    apt-get update

RUN apt-get install -y openssh-server
# Configure SSH
RUN mkdir /var/run/sshd
# 添加公钥
RUN echo 'root:redhat' | chpasswd
RUN echo "PermitRootLogin yes" >> /etc/ssh/sshd_config

# 暴露私钥目录

EXPOSE 22
CMD ["/usr/sbin/sshd", "-D"]
#https://www.runoob.com/w3cnote/set-ssh-login-key.html