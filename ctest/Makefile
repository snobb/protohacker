OBJDIR      := obj
BINDIR      := bin
TARGET      := $(BINDIR)/prog
SRC         := $(wildcard *.c)
OBJ         := $(addprefix $(OBJDIR)/,$(notdir $(SRC:.c=.o)))
CC          ?= cc
CFLAGS      := -Wall
LDFLAGS     :=

ifeq ($(CC), $(filter $(CC), clang gcc cc))
	CFLAGS += -std=c99 -pedantic
endif

all: debug

debug: CFLAGS += -g -DDEBUG
debug: LDFLAGS += -g
debug: build

release: CFLAGS +=-O3
release: clean build
	strip $(TARGET)

build: $(OBJDIR) $(BINDIR) $(BUILD_HOST) $(TARGET)

$(TARGET): $(BUILD_HOST) $(OBJ)
	$(CC) $(LDFLAGS) -o $@ $(OBJ)

$(OBJDIR)/%.o : %.c
	$(CC) $(CFLAGS) -o $@ -c $?

$(OBJDIR):
	@mkdir -p $(OBJDIR)

$(BINDIR):
	@mkdir -p $(BINDIR)

clean:
	-rm -f *.core
	-rm -rf ./$(BINDIR)
	-rm -rf ./$(OBJDIR)

.PHONY : all debug release build run clean objdir
