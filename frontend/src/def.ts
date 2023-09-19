export enum PingType {
    PING = 1,
    MTR = 2
}

export class MtrRow {
    IP = ""
    Sent = 0
    Recv = 0
    Min = 0
    Max = 0
    Loss = 0.0
    Last = 0
}

export function checkSpec(chartData: Array<number>, chartInstance: echarts.ECharts) :number {
    const sum = chartData.reduce((a, b) => a + b, 0)
    const len = chartData.length
    const avg = sum / len
    const diff = chartData.map(v => Math.pow(v - avg, 2)).reduce((a, b) => a + b, 0)
    const stddev = Math.sqrt(diff / len)
    const threshold = avg + 4 * stddev

    const copy = chartData.slice(len <= 100 ? 0 : len - 110, len)
    const over = copy.filter(v => v >= threshold)
    if (over.length) {
        const url = chartInstance.getDataURL()
        const images = document.querySelector('.images')
        const img = document.createElement('img')
        const span = document.createElement('div')
        const maxOver = over.reduce((a, b) => a > b ? a : b, Number.MIN_VALUE)
        span.innerText = `均值: ${avg.toFixed(2)}, 规格: ${threshold.toFixed(2)}, 最大超规: ${maxOver}`

        img.src = url
        images?.appendChild(span)
        images?.appendChild(img)
        return maxOver
    }
    return 0
}