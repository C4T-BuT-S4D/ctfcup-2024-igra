#!/usr/bin/env python3

import enum
import sys

SCREEN_SIZE = 64
field = [[0] * SCREEN_SIZE for _ in range(SCREEN_SIZE)]
player = (0, 0)


class Move(enum.Enum):
    UP = 1
    DOWN = 2
    LEFT = 3
    RIGHT = 4


def read_input() -> set[Move]:
    size = int.from_bytes(sys.stdin.buffer.read(4), "big")

    if size == 0:
        return set()

    keys = sys.stdin.buffer.read(size)

    moves = set()
    for key in keys:
        try:
            moves.add(Move(key))
        except ValueError:
            continue

    return moves


def write_output():
    b = b""
    for row in field:
        b += bytes(row)
    sys.stdout.buffer.write(b)
    sys.stdout.buffer.flush()


def calc_move(pos: tuple[int, int], delta: tuple[int, int]) -> tuple[int, int]:
    target = (pos[0] + delta[0], pos[1] + delta[1])
    if (
        target[0] < 0
        or target[0] >= SCREEN_SIZE
        or target[1] < 0
        or target[1] >= SCREEN_SIZE
    ):
        return pos
    return target


while True:
    delta = (0, 0)
    for move in read_input():
        match move:
            case Move.UP:
                d = (-1, 0)
            case Move.DOWN:
                d = (1, 0)
            case Move.LEFT:
                d = (0, -1)
            case Move.RIGHT:
                d = (0, 1)
        delta = (delta[0] + d[0], delta[1] + d[1])

    next_player = calc_move(player, delta)
    if next_player[0] != player[0] or next_player[1] != player[1]:
        player = next_player
        field[player[0]][player[1]] = (field[player[0]][player[1]] + 1) % 256

    write_output()
