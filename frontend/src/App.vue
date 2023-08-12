<template>
  <div class="root">
    <div class="setting">
      <div>
        <span>目标地址:</span><input type="text" v-model="dest">
      </div>
      <div>
        <span>探测次数:</span><input type="number" v-model="count" :disabled="tillCut">
        <span>持续探测</span><input type="checkbox" v-model="tillCut" />
      </div>
      <div class="item">
        <span>探测间隔（秒）:</span><input type="number" min="0.1" step="0.1" v-model="intervalSec">
      </div>
    </div>
    <div class="switch">
      PING: <input type="checkbox" name="type" id="ping" :checked="isPing" @change="checkBoxChange">
      MTR: <input type="checkbox" name="type" id="mtr" :checked="!isPing" @change="checkBoxChange">
    </div>
    <div class="logo">
      <button class="start" @click="startDetect">开始探测</button>
      <button class="stop" @click="stopDetect">停止探测</button>
    </div>
    <div v-if="detectType === PingType.PING" class="tab">
      <div class="pingRet">
        <span>发送：</span><span class="descValue">{{ pingStat.send }}</span>
        <span>接收：</span><span class="descValue">{{ pingStat.recv }}</span>
        <span>最小：</span><span class="descValue">{{ pingStat.min }} Ms</span>
        <span>平均：</span><span class="descValue">{{ pingStat.avg }} Ms</span>
        <span>最大：</span><span class="descValue">{{ pingStat.max }} Ms</span>
        <span>离散：</span><span class="descValue">{{ pingStat.std }} Ms</span>
        <span>丢包：</span><span class="descValue">{{ pingStat.loss }} %</span>
      </div>
      <div id="chart"></div>
    </div>
    <div v-else class="tab">
      MTR On {{ dest }}
      <table border="1">
        <tr>
          <th>IP</th>
          <th>发送</th>
          <th>接收</th>
          <th>丢包(%)</th>
          <th>最小(ms)</th>
          <th>最大(ms)</th>
          <th>上一个(ms)</th>
        </tr>
        <tr v-for="(row, index) in mtrTable.table" :key="index">
          <td style="text-align: left;">{{ row.IP }}</td>
          <td>{{ row.Sent }}</td>
          <td>{{ row.Recv }}</td>
          <td>{{ parseFloat(row.Loss.toFixed(3)) }}</td>
          <td>{{ row.Min }}</td>
          <td>{{ row.Max }}</td>
          <td>{{ row.Last }}</td>
        </tr>
      </table>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { Mtr, Ping } from '../wailsjs/go/main/App'
import { EventsEmit, EventsOn, EventsOnMultiple } from '../wailsjs/runtime/runtime'
import * as echarts from 'echarts'
import { MtrRow, PingType } from './def';

const detectType = ref<PingType>(PingType.PING)
const isPing = computed(() => detectType.value === PingType.PING)

const dest = ref('www.baidu.com')
const count = ref(50)
const tillCut = ref(true)
const intervalSec = ref(0.2)
const pingStat = reactive({
  recv: 0,
  send: 0,
  loss: 0,
  min: 0,
  max: 0,
  avg: 0,
  std: 0
})

const mtrTable = reactive({
  table: Array<MtrRow>()
})

let chart = null as any
const chartData = reactive({
  xData: Array<string>(),
  yData: Array<number>()
})

async function startDetect() {
  const cnt = tillCut.value ? -1 : count.value
  const interval = intervalSec.value * 1000
  if (isPing.value) {
    ping(cnt, interval)
  } else {
    mtr(cnt, interval)
  }
}


async function ping(cnt: number, interval: number) {
  chartData.xData = []
  chartData.yData = []
  Ping(cnt, interval, dest.value)
}

async function mtr(cnt: number, interval: number) {
  mtrTable.table = []
  Mtr(cnt, interval, dest.value)
}

function stopDetect() {
  if (isPing.value) {
    stopPing()
  } else {
    stopMtr()
  }
}

async function stopPing() {
  EventsEmit("PING_STOP")
}

async function stopMtr() {
  EventsEmit("MTR_STOP")
}

function checkBoxChange(evt: Event) {
  detectType.value = (evt.target as Element)?.id === 'ping' ? PingType.PING : PingType.MTR
}

EventsOn('PING', (d: string) => {
  const ptr = JSON.parse(d)
  chartData.xData.push(ptr.TimeStr as string)
  chartData.yData.push(toMs(ptr.Rtt))
  chart.setOption({
    series: [
      {
        data: chartData.yData
      }
    ],
    xAxis: {
      data: chartData.xData
    }
  })
})

EventsOn("PING_STAT", (d: string) => {
  const stat = JSON.parse(d)
  pingStat.send = stat.PacketsSent
  pingStat.recv = stat.PacketsRecv
  pingStat.loss = parseFloat((stat.PacketLoss as number).toFixed(2))
  pingStat.min = toMs(stat.MinRtt)
  pingStat.avg = toMs(stat.AvgRtt)
  pingStat.max = toMs(stat.MaxRtt)
  pingStat.std = toMs(stat.StdDevRtt)
})

EventsOn("MTR_INIT", (d: string) => {
  const row = JSON.parse(d)
  row.forEach((r: MtrRow) => (r.Min = 0))
  mtrTable.table = row
})

EventsOn("MTR", (d: string) => {
  const row = JSON.parse(d) as MtrRow

  const ri = mtrTable.table.findIndex(r => r.IP === row.IP)
  ri > -1 && (mtrTable.table[ri] = row)
})

const options = {
  visualMap: {
    show: false,
    type: 'continuous',
    seriesIndex: 0,
    min: 0,
    max: 200
  },
  title: {
    left: 'center',
    text: `ping ${dest.value}`
  },
  tooltip: {
    trigger: 'axis',
    formatter: '第 {b} 个 <br> {c} ms',
  },
  grid: {
    left: '5%',
    right: '5%',
    bottom: '20%'
  },
  toolbox: {
    feature: {
      dataZoom: {
        yAxisIndex: false
      },
      restore: {},
      saveAsImage: {},
    }
  },
  xAxis: {
    data: chartData.xData,
    axisLabel: {
      show: true,
      rotate: 90
    },
  },
  yAxis: {
    name: 'ms',
    axisLine: {
      show: true
    }
  },
  series: [
    {
      type: 'line',
      showSymbol: false,
      data: chartData.yData
    }
  ]
}

onMounted(() => {
  chart = echarts.init(document.getElementById('chart'))
  chart.setOption(options)
})

const toMs = (time: any): number => parseFloat(((time as number) / 1000 / 1000).toFixed(2))

</script>

<style>
.root {
  padding: 20px;
}

input {
  margin-left: 8px;
}

#logo {
  display: block;
  width: 50%;
  height: 50%;
  margin: auto;
  padding: 10% 0 0;
  background-position: center;
  background-repeat: no-repeat;
  background-size: 100% 100%;
  background-origin: content-box;
}

.start,
.stop {
  width: 50px;
  height: 50px;
  border-radius: 300px;
  text-align: center;
  cursor: pointer;
  padding: 8px;
  margin: 10px auto;
  border: 1px solid lightskyblue;
  background: lightskyblue;
  font-weight: bold;
  font-size: 14px;

}

.stop {
  margin-left: 5px;
  background: lightgray;
  border-color: lightgray;
}

#chart {
  width: 80%;
  height: 500px;
  padding: 4px;
  border: 1px solid lightgray;
  margin: 10px auto;
}

.setting div {
  display: inline-block;
  margin: 5px 10px;
  padding-right: 10px;
  border-right: 2px solid lightgray;
}

.descValue {
  padding: 4px 10px;
  padding-left: 4px;
  margin-right: 5px;
  font-weight: bold;
  border-right: 2px solid lightgrey;
}

.tab {
  text-align: center;
  padding: 10px 20px;
  margin: auto;
}

table {
  margin: auto;
}
td, th {
  padding: 2px 10px;
  text-align: center;
}
</style>
