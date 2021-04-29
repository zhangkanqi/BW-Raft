# -*- coding: utf-8 -*-
import matplotlib.pyplot as plt

x =  [25, 50, 75, 100, 125, 150, 175, 200, 250]
x_ticks = [str(v) for v in list(x)]
y_Raft =  [290, 980, 1800, 3000, 4600,  5100, 6400, 7100, 6100]
y_BWRaft = [500, 1400, 2900, 4200, 6000,  7700, 8900, 10000, 9000]
y_BFT_BWRaft = [450, 1200, 2600, 3800, 5643,  6709, 7960, 8300, 8000]

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
plt.xlabel("Client Scale", font1)  # X轴标签
plt.ylabel("Write Throughput (ops/sec)", font1)  # Y轴标签
# plt.title("A simple plot") #标题
plt.savefig('./throughput-write.jpg')


