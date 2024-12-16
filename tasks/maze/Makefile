CXX = g++
CXXFLAGS = -static -O0 -Wall -std=c++17
TARGET = maze
SRC = maze.cpp

all: $(TARGET)

$(TARGET): $(SRC)
	$(CXX) $(CXXFLAGS) -o $(TARGET) $(SRC)

clean:
	rm -f $(TARGET)