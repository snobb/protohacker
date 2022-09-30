#include <stdio.h>
#include <stdlib.h>

struct Message {
    char type;
    int32_t time;
    int32_t data;
} __attribute__((packed));

static inline int32_t convert(int32_t data) {
    data = ((data & 0xff) << 24)
           ^ (((data >> 8) & 0xff) << 16)
           ^ (((data >> 16) & 0xff) << 8)
           ^ (((data >> 24) & 0xff));
    return data;
}

static inline struct Message message_from_buf(const char buf[]) {
    struct Message msg = *(struct Message *)buf;
    msg.time = convert(msg.time);
    msg.data = convert(msg.data);

    return msg;
}

int main() {
    const char buf[] = { 0x49, 0x00, 0x00, 0x30, 0x39, 0x00, 0x00, 0x00, 0x65 };
    struct Message msg = message_from_buf(buf);
    printf("%d %d\n", msg.time, msg.data);
}

