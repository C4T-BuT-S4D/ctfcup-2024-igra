#include <iostream>
#include <utility>
#include <set>
#include <cstdint>
#include <vector>
#include <arpa/inet.h>
#include <cstdlib>

#define SCREEN_SIZE 64
#define SPACE 15
#define WALL 0
#define PLAYER 9
#define END 10
#define PATH 8

enum Move {
    UP = 1,
    DOWN = 2,
    LEFT = 3,
    RIGHT = 4
};

const char* MAZE =
"################################################################"
"##       #           #       #   #       #               #     #"
"## ####### ######### # ### # # # # # ##### ### ######### ### # #"
"## #       #         # #   # # #   #     #   # #   #   #   # # #"
"## # ####### ########### ### # ######### ### # # # ### ### ### #"
"## #   #   #   #         #   # #   #   #     #   #   #       # #"
"## ### # # ### # ### ##### ### # # # # ############# ####### # #"
"##     # #     # #   #       # # #   #           # #     # #   #"
"## ##### ####### # ########### # ##### ######### # ##### # ### #"
"##   # #   #     #               #     #S#   # #   #     # #   #"
"#### # ### ####### ##### ############# # # # # ### # ##### # ###"
"## # #   #       # #   # #       #   # #   # #   # # #     #   #"
"## # ### ####### # # # ### ##### # # ####### # # # # # # ##### #"
"## #   # #       # # #   # #   #   #       # # # # # # #   #   #"
"## ### # # # ####### ### # # ############# # # # # # ### # # ###"
"##   #   # # #     #   #   #           #   # # #   #   # # #   #"
"#### ### # ### ### ### ##### ######### # ### # ####### # # ### #"
"##     # #     # #   # # #   #   #     # #   # #   #   # # #   #"
"## ##### ####### ### # # # ##### # ### # # ### # # # ### # # ###"
"## #     # #       #   # #     # # #   # # #   # # # # # #   # #"
"## # ##### # ##### ##### ##### # # ##### # # ### ### # # ##### #"
"## # #       #   #       #       #     #   #     #   # #   #   #"
"## # # ######### ##### ### ########### ######### # ### ### # # #"
"## # # #   #   #   # #         #     #   #   #   # #     # # # #"
"## # # # # # # # # # ########### ### # # # # ##### ### ### # ###"
"##   #   #   #   #       #       # #   # # #   #   #   #   #   #"
"## ################# ##### ####### ##### # ### # ### ### ##### #"
"## #             #   #     #   #     #   # # #   #       #     #"
"## # ##### ### ### ### ####### # ### # ### # ##### ####### # ###"
"## # #   #   #       #   #     # #   # #   #     #   #   # #   #"
"## ### # ### ########### # ### # # ### # ### ### ### # ### ### #"
"## #   #   #   #         # #   # #   # #   #   #     #     #   #"
"## # ##### ##### ######### ##### ### # # # ######### ### ### # #"
"##   #   #     #       #   #   # #   # # # #       #   # #   # #"
"###### ####### ##### # ### # # # # # # ### # ##### ### ### ### #"
"##           # #   # #   #   #   # # #   # # # #   #       #   #"
"## ####### ### # # ##### # ####### # ### # # # # ####### #######"
"##       # #   # #     # # #       # # # #   # #       # #     #"
"######## # # ### ##### # # # ####### # # # ### ####### ### ### #"
"## #     #   #       #   # # #     #   # #           #     #   #"
"## # ############### ##### # # ### ### # ####### ##### ##### ###"
"## #   #             #   # #     #   # #     #   #   # #   #   #"
"## ### # ##### # ##### # ######### # # ##### ##### # # # # ### #"
"##     #     # # #     #   #     #E# # #   # #     # # # # #   #"
"## ######### # ### ####### # ### ### ### # # # ##### ### # # ###"
"## #     # # # #   #       #   #   #     # # #   # #   # # #   #"
"## # ### # # # # ### ### ##### ### ### ### # ### # ### # # # # #"
"##   #   #   #     # #   #       #   # #   #   # #   #   # # # #"
"###### ##### ####### # ### ######### ### # ### # # # ##### ### #"
"##   # #   # #       # #   #   #     #   #   # # # #     # #   #"
"## ### # # ### # ####### ### # # ##### ##### # # # # ##### # # #"
"##   # # #     # #       #   #   # #     #   #   # #       # # #"
"## # # # ######### ####### ####### # ### ######### ######### # #"
"## # # # #   #   # # #   # #       #   #         # #         # #"
"## # # # # # # # # # # # # # ### # ### ######### # # ##### ### #"
"## #   #   #   # #   # #   #   # # # # #       # # # #   # #   #"
"## ######### ### ### # ######### # # # ##### # # # # # # # # ###"
"##   #     #   # #   #   #       #   #     # # #   #   # # # # #"
"#### # ### ##### # ##### # ### ########### ### ########### # # #"
"## #   # #     #   #     #   # #       #   #     #     #   # # #"
"## ##### ##### ##### ##### # ### ##### # ### ### ### # # ### # #"
"##                   #     #         #       #       #   #     #"
"################################################################"
"################################################################";

int screen[SCREEN_SIZE][SCREEN_SIZE];

void print_screen() {
    for (int i = 0; i < SCREEN_SIZE; ++i) {
        for (int j = 0; j < SCREEN_SIZE; ++j) {
            std::cout << static_cast<char>(screen[i][j]);
        }
    }
    std::cout << std::flush;
}


int main() {
    std::pair<int, int> start;
    std::pair<int, int> end;
    for (int i = 0; i < SCREEN_SIZE; ++i) {
        for (int j = 0; j < SCREEN_SIZE; ++j) {
            char c = MAZE[i * SCREEN_SIZE + j];
            if (c == '#') {
                screen[i][j] = SPACE;
            } else if (c == 'S') {
                start = std::make_pair(j, i);
            } else if (c == 'E') {
                screen[i][j] = END;
                end = std::make_pair(j, i);
            } else if (c == '.') {
                screen[i][j] = SPACE;
            } else {
                screen[i][j] = SPACE;
            }
            
        }
    }
    std::pair<int, int> player = start;
    screen[start.second][start.first] = PLAYER;

    bool won = false;
    bool lost = false;

    std::set<std::pair<int, int>> visited;
    std::pair<int, int> delta = std::make_pair(0, 0);
    std::set<Move> moves;
    while (true) {
        if (won) {
            std::cout << "WIN" << std::flush;
            continue;
        }
        if (lost) {
            std::cout << "LOSE" << std::flush;
            continue;
        }

        delta.first = 0;
        delta.second = 0;
        moves.clear();

        uint32_t size;
        if (!std::cin.read(reinterpret_cast<char*>(&size), sizeof(size))) {
            throw std::runtime_error("Failed to read size");
        }
        size = ntohl(size);

        std::cerr << "Size: " << size << std::endl;
        std::vector<char> buffer(size);

        if (size != 0) {
            std::cin.read(buffer.data(), size);
        }

        for (char c : buffer) {
            moves.insert(static_cast<Move>(c));
        }

        for (Move move : moves) {
            std::cerr << "Move: " << move << std::endl;
            switch (move) {
                case UP:
                    delta.second = -1;
                    break;
                case DOWN:
                    delta.second = 1;
                    break;
                case LEFT:
                    delta.first = -1;
                    break;
                case RIGHT:
                    delta.first = 1;
                    break;
            }

            std::pair<int, int> new_pos = std::make_pair(player.first + delta.first, player.second + delta.second);
            // int new_x = player.first + delta.first;
            // int new_y = player.second + delta.second;
            if (new_pos == end) {
                won = true;
            }

            char cell = MAZE[new_pos.second * SCREEN_SIZE + new_pos.first];
            if (cell != '#') {
                screen[player.second][player.first] = SPACE;
                player = new_pos;
                if (visited.find(player) != visited.end()) {
                    lost = true;
                }
                visited.insert(player);
            }

            screen[player.second][player.first] = PLAYER;
        }
        print_screen();
    }

    return 0;
}