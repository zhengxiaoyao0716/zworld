#!/usr/bin/python
# -*- coding: utf-8 -*-

"""
球体建模尝试与验证
"""

import sys
import math
import random
import abc

from sortedcontainers import SortedSet
from pylab import mpl
import matplotlib.pyplot as plt
import numpy as np
from mpl_toolkits.mplot3d import Axes3D as _

range = xrange


def draw_sphere(ax, alpha=0.5):
    """绘制球体"""
    ax.set_aspect('equal')
    u = np.linspace(0, 2 * np.pi, 100)
    v = np.linspace(0, np.pi, 100)
    x = np.outer(np.cos(u), np.sin(v))  # x = R*CosU*SinV
    y = np.outer(np.sin(u), np.sin(v))  # y = R*SinU*SinV
    z = np.outer(np.ones(np.size(u)), np.cos(v))  # z = R*CosV
    ax.plot_surface(x, y, z,  rstride=5, cstride=5,
                    color='white', linewidth=0, alpha=alpha)

    ax.plot(np.zeros(np.size(u)), np.sin(u), np.cos(u),
            color='r', linestyle='dashed', alpha=alpha)
    ax.plot(np.sin(u), np.zeros(np.size(u)), np.cos(u),
            color='g', linestyle='dashed', alpha=alpha)
    ax.plot(np.sin(u), np.cos(u), np.zeros(np.size(u)),
            color='b', linestyle='dashed', alpha=alpha)


def rand_point():
    """随机坐标点"""
    rp = 1.0
    x = 2 * (random.random() - 0.5)
    rp -= math.pow(x, 2)
    y = 2 * math.sqrt(rp) * (random.random() - 0.5)
    rp -= math.pow(y, 2)
    z = math.sqrt(rp) * (random.randint(0, 1) * 2 - 1)
    return x, y, z


class Point(object):
    def __init__(self, index=None, coord=None):
        self._index = index
        self._coord = coord

    @property
    def index(self):
        if not isinstance(self._index, int):
            self._index = self._index()
        return self._index

    @property
    def coord(self):
        if not isinstance(self._coord, tuple):
            self._coord = self._coord()
        return self._coord


class PointSet(object):
    """点集"""
    __metaclass__ = abc.ABCMeta

    def __init__(self, n=1000):
        self.n = n

    @abc.abstractmethod
    def index(self, z):
        """取某个坐标（附近）的索引"""
        pass

    @abc.abstractmethod
    def coord(self, i):
        """取某个索引的（准确）坐标"""
        pass

    def point(self, *payload):
        if len(payload) == 1:
            i, = payload
            return Point(lambda: i, lambda: self.coord(i))
        _, _, z = payload
        return Point(lambda: self.index(z), lambda: payload)

    def angle(self, x, y, z):
        """计算某向量的角度"""
        # TODO 计算角度，然后确定划分多边形等等
        pass

    def near(self, payload, n=1):
        """查找离某坐标或样点最近的样点"""
        if isinstance(payload, Point):
            x, y, z = payload.coord
            index = payload.index
            rs = SortedSet([(None, 4.0)], key=lambda v: v[1])
        else:
            x, y, z = payload
            index = self.index(z)
            point = self.point(index)
            rs = SortedSet([(point, reduce(lambda s, i: s + math.pow(point.coord[i] - payload[i], 2), range(3), 0))], key=lambda v: v[1])

        for incre in (-1, 1):
            i = index
            while True:
                i += incre
                if i < 0 or i >= self.n:
                    break
                xi, yi, zi = self.coord(i)

                dz = math.pow(zi - z, 2)
                rd = rs[n-1 if len(rs) >= n else -1][1]
                if dz > rd:
                    break

                dist = math.pow(xi - x, 2) + math.pow(yi - y, 2) + dz
                if dist > rd:
                    continue

                rs.add((self.point(i), dist))
        return rs[0:n]

    def area(self, point):
        """查找离某样点最近的区域"""
        x, y, z = point.coord
        for pi, nd in self.near(point, n=8):
            xn, yn, zn = pi.coord
            ca = (x + xn) / 2, (y + yn) / 2, (z + zn) / 2
            ai, ad = self.near(ca)[0]
            if ai.index == pi.index or ai.index == point.index:
                yield ca

    def each(self):
        """遍历集合"""
        for i in range(0, self.n):
            yield self.point(i)


class Samples(PointSet):
    """样点集合"""
    incre = 2 * math.pi * (math.sqrt(5) - 1) / 2

    def index(self, z):
        return int(((z + 1) * self.n - 1) / 2)

    def coord(self, i):
        z = float(2 * i + 1) / self.n - 1
        rad = math.sqrt(1 - math.pow(z, 2))
        ang = i * self.incre
        x = rad * math.cos(ang)
        y = rad * math.sin(ang)
        return x, y, z


def main():
    """Entrypoint"""
    ax = fig.add_subplot(111, projection='3d')
    plt.title(u'球面坐标建模')
    draw_sphere(ax, 0.1)

    samples = Samples(300)
    # 均匀采样
    cmap = plt.get_cmap("RdYlGn")
    for point in samples.each():
        x, y, z = point.coord
        c = cmap(0.5 + z / 2)
        ax.scatter(x, y, z, c=c)
    # 随机点最近点
    for x, y, z in (rand_point() for _ in range(10)):
        ax.scatter(x, y, z, c='grey')
        xn, yn, zn = samples.near((x, y, z))[0][0].coord
        ax.plot((x, xn), (y, yn), (z, zn), color='grey')
    # 样本点管辖域
    for i in (random.randint(0, samples.n - 1) for _ in range(10)):
        x, y, z = samples.point(i).coord
        ax.scatter(x, y, z, c='black')
        for xa, ya, za in samples.area(samples.point(i)):
            ax.scatter(xa, ya, za, color='black')


if __name__ == '__main__':
    mpl.rcParams['font.sans-serif'] = ['SimHei']
    mpl.rcParams['axes.unicode_minus'] = False
    fig = plt.figure(figsize=(12, 9), dpi=100)
    main()
    sys.stdout.flush()
    plt.show()
