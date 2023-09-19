<template>
  <div class="root">
    <div class="setting">
      <div>
        <span>目标地址:</span><input type="text" v-model="dest" :disabled="detecting">
      </div>
      <div>
        <span>探测次数:</span><input type="number" v-model="count" :disabled="tillCut || detecting" style="width: 40px;">
        <span>持续探测</span><input type="checkbox" v-model="tillCut" :disabled="detecting" />
      </div>
      <div class="item">
        <span>探测间隔（秒）:</span><input type="number" min="0.1" step="0.1" v-model="intervalSec" :disabled="detecting"
          style="width: 40px;">
      </div>
      <div class="item" v-if="!isPing">
        <span>最大跳数:</span><input type="number" min="1" max="10" step="1" v-model="maxHops" :disabled="detecting"
          style="width: 40px;">
      </div>
    </div>
    <div class="switch">
      <button id="ping" :class="pingClassList" :disabled="detecting" @click="checkBoxChange">PING</button>
      <button id="mtr" :class="mtrClassList" :disabled="detecting" @click="checkBoxChange">MTR</button>
    </div>
    <div class="logo">
      <button class="start" @click="startDetect" :disabled="detecting">开始</button>
      <button class="stop" @click="stopDetect">停止</button>
    </div>
    <div v-show="isPing" class="tab">
      <div class="loading" v-if="detecting">{{ detectingText }}</div>
      <div class="pingRet">
        <span class="descValue">{{ chartData.yData.length }}</span>
        <span>发送：</span><span class="descValue">{{ pingStat.send }}</span>
        <span>接收：</span><span class="descValue">{{ pingStat.recv }}</span>
        <span>最小：</span><span class="descValue">{{ pingStat.min }} Ms</span>
        <span>平均：</span><span class="descValue">{{ pingStat.avg }} Ms</span>
        <span>最大：</span><span class="descValue">{{ pingStat.max }} Ms</span>
        <span>离散：</span><span class="descValue">{{ pingStat.std }} Ms</span>
        <span>丢包：</span><span class="descValue">{{ pingStat.loss }} %</span>
      </div>
      <div id="chart"></div>
      <div id="violator">异常值截图数：{{ violatorCount }}</div>
      <div class="images"></div>
    </div>
    <div v-show="!isPing" class="tab">
      MTR: {{ dest }}
      <div class="loading" v-if="detecting">{{ detectingText }}</div>
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
import { EventsEmit, EventsOn } from '../wailsjs/runtime/runtime'
import * as echarts from 'echarts'
import { checkSpec, MtrRow, PingType } from './def';

const detectType = ref<PingType>(PingType.PING)
const isPing = computed(() => detectType.value === PingType.PING)
const detecting = ref(false)
const detectingText = computed(() => detecting.value ? '探测中...' : '')
const pingClassList = computed(() => detectType.value === PingType.PING ? 'optionBtn active' : 'optionBtn')
const mtrClassList = computed(() => detectType.value === PingType.MTR ? 'optionBtn active' : 'optionBtn')

const dest = ref('www.baidu.com')
const count = ref(50)
const tillCut = ref(true)
const intervalSec = ref(0.2)
const maxHops = ref(4)
const pingStat = reactive({
  recv: 0,
  send: 0,
  loss: 0,
  min: 0,
  max: 0,
  avg: 0,
  std: 0
})
const checkCount = ref(0)
const violatorCount = ref(0)

const mtrTable = reactive({
  table: Array<MtrRow>()
})

let chart: echarts.ECharts
const chartData = reactive({
  xData: Array<string>(),
  yData: Array<number>()
})

async function startDetect() {
  const cnt = tillCut.value ? -1 : count.value
  const interval = intervalSec.value * 1000
  detecting.value = true
  if (isPing.value) {
    ping(cnt, interval)
  } else {
    mtr(cnt, interval)
  }
}


async function ping(cnt: number, interval: number) {
  chartData.xData = []
  chartData.yData = []
  pingStat.send = 0
  pingStat.recv = 0
  pingStat.loss = 0
  pingStat.min = 0
  pingStat.avg = 0
  pingStat.max = 0
  pingStat.std = 0

  const images = document.querySelector('.images')
  Array.from(images?.children || []).forEach(child => images?.removeChild(child))
  violatorCount.value = 0

  Ping(cnt, interval, dest.value)
}

async function mtr(cnt: number, interval: number) {
  mtrTable.table = []
  Mtr(cnt, interval, maxHops.value, dest.value)
}

function stopDetect() {
  detecting.value = false
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
  checkCount.value++
  const ptr = JSON.parse(d)
  chartData.xData.push(ptr.TimeStr as string)
  chartData.yData.push(toMs(ptr.Rtt))
  const option = {
    title: {
      left: 'center',
      text: `ping ${dest.value}`
    },
    series: [
      {
        data: chartData.yData
      }
    ],
    xAxis: {
      data: chartData.xData
    }
  } as echarts.EChartsOption
  if (chartData.yData.length > 1000) {
    option.animation = false;
    (option.series as Array<echarts.LineSeriesOption>)[0].sampling = 'max'
  }
  chart.setOption(option)
  if (checkCount.value !== 0 && checkCount.value % 100 === 0) {
    const over = checkSpec(chartData.yData, chart)
    if (over) {
      violatorCount.value++
    }
    checkCount.value = 0
  }
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

  detecting.value = false
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
    formatter: '📌{b} <br> ⏰{c} ms',
    className: 'tooltip'
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
} as echarts.EChartsOption

onMounted(() => {
  chart = echarts.init(document.getElementById('chart'), null, { renderer: 'svg' })
  chart.setOption(options)
})

const toMs = (time: any): number => parseFloat(((time as number) / 1000 / 1000).toFixed(2))

</script>

<style>
html {
  --borderRadius: 5px
}

.root {
  padding: 20px;
}

.root button:disabled,
.root input:disabled {
  cursor: not-allowed;
  background: lightgray;
}

.setting {
  border: 1px solid lightgray;
  padding: 5px 10px;
  margin: 5px;
  border-radius: 15px;
}

input {
  margin-left: 8px;
  border-radius: var(--borderRadius);
  border: 1px solid lightgray;
  margin-right: 3px;
}

.optionBtn {
  line-height: 20px;
  width: 300px;
  font-size: 15px;
  border: 1px solid lightgray;
  border-radius: var(--borderRadius);
  margin-left: 5px;
}

.active {
  line-height: 25px;
  font-weight: bold;
  font-size: 18px;
  background: lightskyblue;
}

.loading {
  font-weight: bold;
  width: 150px;
  height: 25px;
  padding: 2px 12px;
  user-select: none;
  margin: 4px auto;
  border: 1px solid lightgray;
  border-radius: var(--borderRadius);
  animation-name: loading-s;
  animation-duration: 1s;
  animation-fill-mode: backwards;
  animation-iteration-count: infinite;
  animation-direction: alternate;
}

@keyframes loading-s {
  0% {
    font-size: 12px;
    color: gray;
  }

  100% {
    font-size: 18px;
    color: black;
    padding: 2px 16px;
    border-radius: 10px;
  }
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
  width: 95%;
  height: 500px;
  padding: 4px;
  border: 1px solid lightgray;
  margin: 10px auto;
  border-radius: var(--borderRadius);
}

.images {
  margin: 15px 5px;
  padding: 5px;
  border: 2px solid lightblue;
  border-radius: var(--borderRadius);
  height: 700px;
  overflow-y: scroll;
}

.images>img {
  margin: auto;
  margin-bottom: 5px;
  border: 1px solid lightcoral;
  border-radius: var(--borderRadius);
}

#violator {
  color: red;
  font-weight: bold;
}

.setting div {
  display: inline-block;
  margin: 5px 10px;
  padding-right: 10px;
  border-right: 2px solid lightgray;
}

.descValue {
  padding: 4px 5px;
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

td,
th {
  padding: 2px 10px;
  text-align: center;
}

.tooltip {
  --c: red;
  border-radius: 100px;
  background: var(--c);
  margin-right: 2px;
  color: var(--c);
  text-align: left;
}
</style>
