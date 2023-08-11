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
    <div class="logo">
      <button class="start" @click="ping">开始探测</button>
      <button class="stop" @click="stopPing">停止探测</button>
    </div>
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
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue'
import { Ping } from '../wailsjs/go/main/App'
import { EventsEmit, EventsOn } from '../wailsjs/runtime/runtime'
import * as echarts from 'echarts'

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

let chart = null as any
const chartData = reactive({
  xData: Array<string>(),
  yData: Array<number>()
})


async function ping() {
  chartData.xData = []
  chartData.yData = []
  const cnt = tillCut.value ? -1 : count.value
  Ping(cnt, intervalSec.value * 1000, dest.value)
}

async function stopPing() {
  EventsEmit("PING_STOP")
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
</style>
