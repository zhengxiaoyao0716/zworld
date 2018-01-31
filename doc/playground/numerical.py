#!/usr/bin/python
# -*- coding: utf-8 -*-

"""
一些数值验证
"""

from functools import reduce
import sys

from pylab import mpl
import matplotlib.pyplot as plt
import numpy as np


def plot_area():
    """拟合面积特征"""
    # 亚洲 非洲 北美 南美 南极 欧洲 大洋洲
    area = tuple(a / 100 for a in (29.4, 20.2, 16.2, 12.0, 9.4, 6.8, 6.0))
    print('area:', area)

    avg = reduce(lambda l, r: l + r, area) / len(area)
    diff = tuple(abs(a - avg) for a in area)
    print('diff:', diff)

    x = np.arange(0, 1, 1.0 / len(area))
    z = np.polyfit(x, area, 2)
    p = np.poly1d(z)
    ps = '(%fx^2) + (%fx) + (%f)' % (p[2], p[1], p[0])
    pval = p(x)
    print('poly:', ps)

    plt.plot(x, area, 'xr', label=u'面积占总陆地面积比')
    plt.plot(x, diff, 'og', label=u'与平均值差的绝对值')
    plt.plot(x, pval, 'b', label=u'拟合函数（2次）')
    plt.text(x[1], pval[1], ps)

    hight = tuple(float(h) / 10000 for h in (950,
                  650, 700, 600, 2350, 300, 400))
    plt.plot(x, hight, 'black', label=u'平均海拔 / 10000')


def plot_hight(level):
    """拟合高度特征"""
    hight = (0, 850, 8844)
    x = (level, (1 + level) / 2, 1)
    plt.plot(x, hight, 'r')

    x = np.arange(level, 1.01, (1 - level) / 20)
    p = lambda x: 9000 * ((x - level) / (1 - level)) ** 3.5
    pval = tuple(p(xi) for xi in x)
    plt.plot(x, pval, 'b', label=u'拟合函数（3次）')


def main():
    """Entrypoint"""
    plt.subplot(221)
    plt.title(u'各大洲面积特征')
    plt.xlabel(u'index / 7')
    plt.ylabel(u'area / total')
    plot_area()
    plt.legend(loc=4)

    plt.subplot(222)
    plt.title(u'陆地（高于水平面）高度特征')
    plt.xlabel(u'random')
    plt.ylabel(u'hight')
    plot_hight(0.7)
    plt.legend(loc=4)

if __name__ == '__main__':
    mpl.rcParams['font.sans-serif']=['SimHei']
    plt.figure(figsize=(16, 9), dpi=100)
    main()
    sys.stdout.flush()
    plt.show()
