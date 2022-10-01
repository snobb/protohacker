#include <stdio.h>
#include <stdlib.h>
#include <arpa/inet.h>

struct Message {
    char type;
    int32_t time;
    int32_t data;
} __attribute__((packed));

static inline struct Message message_from_buf(const char buf[]) {
    struct Message msg = *(struct Message *)buf;
    msg.time = ntohl(msg.time);
    msg.data = ntohl(msg.data);

    return msg;
}

int main() {
    const char buf[] = { 0x49, 0x00, 0x00, 0x30, 0x39, 0x00, 0x00, 0x00, 0x65 };
    struct Message msg = message_from_buf(buf);
    printf("%d %d\n", msg.time, msg.data);
}

