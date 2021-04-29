# -*- coding: utf-8 -*-
import matplotlib.pyplot as plt

x =  [5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31]
x_ticks = [str(v) for v in list(x)]
y_Raft =  [928, 1423, 2250, 3152, 3843,  5389, 6478, 7821, 8589, 9645, 11784, 13451, 15949, 18495]
y_BWRaft = [900, 1323, 2000, 3404, 4243,  5689, 6540, 8371, 9189, 10045, 11584, 13851, 16449, 18995]
y_BFT_BWRaft = [1090, 1923, 3600, 5104, 5943,  7009, 8600, 9991, 11009, 14045, 18584, 20851, 26449, 32995]

# plt.xlim(20, 260)  # 限定横轴的范围
# plt.ylim(350, 14000)  # 限定纵轴的范围

font1 = {'family': '',
'weight': 'normal',
'size': 17,
}

plt.figure(figsize=(10, 6), dpi=500)

plt.plot(x, y_Raft, marker='*', linewidth=2.7, ms=8, label='Raft',)
plt.plot(x, y_BWRaft, marker='o', linewidth=2.7, ms=8, label='BW-Raft')
plt.plot(x, y_BFT_BWRaft, marker='d', linewidth=2.7, ms=8, label='BFT_BW-Raft')

plt.legend(prop=font1, framealpha=0.2, loc='upper left')  # 让图例生效

plt.xticks(x, x_ticks, rotation=1)

plt.margins(0)
plt.subplots_adjust(bottom=0.2)
plt.xlabel("Cluster Size", font1)  # X轴标签
plt.ylabel("Leader Election Time (ms)", font1)  # Y轴标签
# plt.title("A simple plot") #标题
plt.savefig('./electTime.jpg')


