#include <iostream>
#include <cstdlib>
#include <sys/select.h>
#include <sys/socket.h>
#include <unistd.h>
#include <stdio.h>  
#include <sys/types.h>  
#include <netinet/in.h>  
#include <arpa/inet.h>
#include <string>
#include <string.h>

/*
 g++ client.cpp -o client
*/

#define BUFFER_LENGTH 1024

int main(int argc, char *argv[])
{
    int client_fd = socket(AF_INET, SOCK_STREAM, 0);
    if(client_fd < 0){
        printf("create socket fd error");
        exit(0);
    }

    struct sockaddr_in server_addr;
    memset(&server_addr, 0 , sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_addr.s_addr = inet_addr("127.0.0.1"); 
    server_addr.sin_port = htons(8888);

    int ret = connect(client_fd, (struct sockaddr*)&server_addr, sizeof(sockaddr_in));
    if(ret < 0 ){
        printf("connect error");
        exit(0);
    }
    char szBuf[BUFFER_LENGTH] = {0};
    while(true)
    {   
        printf("sys>");
        memset(szBuf, 0 , sizeof(szBuf));
        //gets(szBuf);
        scanf("%[^\n]", &szBuf);
        if (strcmp("exit", szBuf) == 0){
            close(client_fd);
            break;
        }
        send(client_fd, szBuf, strlen(szBuf), 0);

        memset(szBuf, 0 , sizeof(szBuf));
        recv(client_fd, szBuf, BUFFER_LENGTH, 0);
        printf("ack: %s\n",szBuf);


    }

}