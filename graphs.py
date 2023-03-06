# csv format:
# transaction_id, time (nanoseconds)

import pandas as pd
import matplotlib.pyplot as plt
import numpy as np

NUM_FILES = 3

DATA_FILE_NAME = "logs/node{}_transactions_log.csv"


allDfs = []

for i in range(1, NUM_FILES+1):
    df = pd.read_csv(DATA_FILE_NAME.format(i))
    # print(df['time'])
    allDfs.append(df)

df = pd.concat(allDfs, axis=0)

df['time'] = df['time'].apply(
    lambda time: float(time/(10**6)))

print(df.loc[df['transaction_id'] == 'node2_1024_T'])

times_series = df.groupby('transaction_id')['time'].apply(lambda g: g.max() - g.min())

print("LARGEST: \n", times_series.nlargest(60))
print("SMALLEST: \n", times_series.nsmallest(60))
times_arr = times_series.to_numpy()

x = np.sort(times_arr)
y = 100 * np.arange(1, len(x) + 1) / len(x)

print(x, y)
# plt.plot(x, y,  marker=".", linestyle="none")
plt.plot(x, y)

plt.xlabel('Transaction Processing Time (ms)')
plt.ylabel('Percentile (%)')
plt.show()
