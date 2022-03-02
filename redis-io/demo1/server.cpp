#include <iostream>
#include <cstdlib>
#include <unistd.h>
#include <stdio.h>  
#include <sys/types.h>  
#include <sys/select.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <string>
#include <string.h>
#include <errno.h>

#define BUFFER_LENGTH	1024
#define POLL_SIZE		1024
#define EPOLL_SIZE		1024

int main(int argc, char *argv[])
{
    int listen_fd = socket(AF_INET, SOCK_STREAM, 0);
    if (listen_fd < 0){
        printf("create socket error %s %d\n", strerror(errno), errno);
        return -1;
    }

    struct sockaddr_in addr;
    memset(&addr, 0 , sizeof(addr));
    addr.sin_family = AF_INET;
    addr.sin_addr.s_addr = htonl(INADDR_ANY);
    addr.sin_port = htons(8888);
    if (bind(listen_fd, (struct sockaddr*)&addr, sizeof(addr)) < 0){
        printf("bind socket error, %s %d\n", strerror(errno), errno);
        exit(0);
    }

    if (listen(listen_fd, 10) < 0){
        printf("listen socket error, %s %d\n", strerror(errno), errno);
        exit(0);
    }

    struct sockaddr_in client_addr;
	memset(&client_addr, 0, sizeof(struct sockaddr_in));
	socklen_t client_len = sizeof(client_addr);
    int clientfd = accept(listen_fd, (struct sockaddr*)&client_addr, &client_len);
    if (clientfd < 0){
        printf("accect error");
        exit(0);
    }
    printf("new client success\n");

    char buffer[BUFFER_LENGTH] = {0};
    int ret = 0;
    while(true)
    {
        memset(buffer, 0 , sizeof(buffer));
        ret = recv(clientfd, buffer, BUFFER_LENGTH, 0);
        if ( ret < 0){
            if (errno == EAGAIN || errno == EWOULDBLOCK) {
				printf("read all data\n");
			}
			break;
        }
        else if ( ret == 0){
            printf("disconnect \n");
			break;
        }
        else {
            printf("recv: %s\n", buffer);
            send(clientfd, buffer, strlen(buffer), 0);
        }
    }
    close(clientfd);
    close(listen_fd);
    return 0;
}