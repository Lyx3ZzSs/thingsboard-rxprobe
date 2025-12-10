<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const props = defineProps({
  title: {
    type: String,
    required: true
  },
  data: {
    type: Array,
    default: () => []
  },
  labels: {
    type: Array,
    default: () => []
  },
  color: {
    type: String,
    default: '#0ea5e9' // 电力蓝
  },
  unit: {
    type: String,
    default: ''
  },
  fill: {
    type: Boolean,
    default: true
  },
  height: {
    type: Number,
    default: 180
  }
})

const chartData = computed(() => ({
  labels: props.labels,
  datasets: [
    {
      label: props.title,
      data: props.data,
      borderColor: props.color,
      backgroundColor: props.fill 
        ? `${props.color}15`
        : 'transparent',
      borderWidth: 2,
      fill: props.fill,
      tension: 0.4,
      pointRadius: 0,
      pointHoverRadius: 4,
      pointHoverBackgroundColor: props.color,
      pointHoverBorderColor: 'hsl(220, 25%, 8%)',
      pointHoverBorderWidth: 2
    }
  ]
}))

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index'
  },
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      backgroundColor: 'hsl(220, 25%, 10%)',
      borderColor: 'hsl(199, 89%, 48%, 0.3)',
      borderWidth: 1,
      titleColor: 'hsl(200, 20%, 95%)',
      bodyColor: 'hsl(215, 15%, 65%)',
      padding: 10,
      cornerRadius: 6,
      displayColors: false,
      callbacks: {
        label: (context) => {
          return `${context.parsed.y.toFixed(1)}${props.unit}`
        }
      }
    }
  },
  scales: {
    x: {
      display: true,
      grid: {
        display: false
      },
      ticks: {
        color: 'hsl(215, 15%, 45%)',
        maxRotation: 0,
        maxTicksLimit: 6,
        font: {
          size: 10
        }
      },
      border: {
        display: false
      }
    },
    y: {
      display: true,
      grid: {
        color: 'hsl(215, 25%, 15%)',
        drawBorder: false
      },
      ticks: {
        color: 'hsl(215, 15%, 45%)',
        callback: (value) => `${value}${props.unit}`,
        font: {
          size: 10
        }
      },
      border: {
        display: false
      }
    }
  }
}))
</script>

<template>
  <Card class="overflow-hidden border-primary/20">
    <CardHeader class="pb-2">
      <CardTitle class="text-sm font-medium text-muted-foreground">{{ title }}</CardTitle>
    </CardHeader>
    <CardContent class="pt-0">
      <div :style="{ height: `${height}px` }">
        <Line :data="chartData" :options="chartOptions" />
      </div>
    </CardContent>
  </Card>
</template>
