#!/usr/bin/env python3

import copy
import enum
import random
import sys

rnd = random.SystemRandom()

SCREEN_SIZE = 64
field = [[0] * SCREEN_SIZE for _ in range(SCREEN_SIZE)]
player = (0, 0)


def gen_location() -> tuple[int, int]:
    return (rnd.randint(0, SCREEN_SIZE - 1), rnd.randint(0, SCREEN_SIZE - 1))


target = gen_location()
enemies = [gen_location() for _ in range(32)]


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


def write_output(screen: list[list[int]]):
    b = b""
    for row in screen:
        b += bytes(row)
    sys.stdout.buffer.write(b)
    sys.stdout.buffer.flush()


def gen_delta() -> tuple[int, int]:
    return (rnd.randint(-1, 1), rnd.randint(-1, 1))


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


def is_enemy_hit() -> bool:
    return any(player == enemy for enemy in enemies)


def caught_target() -> bool:
    return player == target


def win():
    out = copy.deepcopy(field)
    out[0][0:3] = b"WON"
    write_output(out)
    sys.exit(0)


def lose():
    out = copy.deepcopy(field)
    out[0][0:4] = b"LOSE"
    write_output(out)
    sys.exit(0)


lost, won = False, False

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

    if lost:
        lose()
    elif won:
        win()

    next_player = calc_move(player, delta)
    if next_player != player:
        player = next_player
        if is_enemy_hit():
            lost = True
        field[player[0]][player[1]] = (field[player[0]][player[1]] + 1) % 256

    for i, enemy in enumerate(enemies):
        enemies[i] = calc_move(enemy, gen_delta())
    if is_enemy_hit():
        lost = True

    target = calc_move(target, gen_delta())
    if caught_target():
        won = True

    screen = copy.deepcopy(field)
    screen[target[0]][target[1]] = 118
    for enemy in enemies:
        screen[enemy[0]][enemy[1]] = 160
    write_output(screen)
