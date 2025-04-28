<script setup lang="ts">
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
// import type { PropType } from 'vue'
import { ref, onMounted } from 'vue'
import Button from 'primevue/button'
// import type { AuthState } from '@/types/auth_state'
import axios from 'axios'
import ProgressBar from 'primevue/progressbar'
import Card from 'primevue/card'
import Divider from 'primevue/divider'
import InputGroup from 'primevue/inputgroup'
import InputText from 'primevue/inputtext'
import DatePicker from 'primevue/datepicker'
import { useToast } from 'primevue/usetoast'

const toast = useToast()

const props = defineProps(['token'])

interface FlightInfo {
  agent: string
  duration: string
  price: string
}

interface SearchResponse {
  cheapest: FlightInfo
  comparison: FlightInfo[]
  fastest: FlightInfo
}

const flights = ref()
const cheapest = ref()
const fastest = ref()
const loading = ref()

const origin = ref()
const destination = ref()
const date = ref()

onMounted(() => {
  console.info('props:')
  console.log(props.token)
})

const formatDate = (d) => {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0') // getMonth() returns 0-11, so add 1
  const day = String(d.getDate()).padStart(2, '0')

  return `${year}-${month}-${day}`
}

const showErr = (msg) => {
  toast.add({
    severity: 'error',
    summary: `Invalid search`,
    detail: msg,
    life: 5000,
  })
}

const onSearch = async () => {
  if (!origin.value) {
    showErr('Origin field is required')
    return
  }

  if (origin.value.length !== 3) {
    showErr('Origin must contain a valid three letter code')
    return
  }

  if (!destination.value) {
    showErr('Destination field is required')
    return
  }

  if (destination.value.length !== 3) {
    showErr('Destination must contain a valid three letter code')
    return
  }

  if (!date.value) {
    showErr('Date is required')
    return
  }

  try {
    loading.value = true
    const d = formatDate(date.value)

    const url = `https://localhost:8443/flights/search?origin=${origin.value}&destination=${destination.value}&date=${d}`

    const response = await axios.get<SearchResponse>(url, {
      headers: {
        'Content-Type': 'application/json',
        Authorization: 'Bearer ' + props.token,
      },
    })

    loading.value = false

    if (Object.keys(response.data).length == 0) {
      toast.add({
        severity: 'warn',
        summary: 'No results',
        detail: "We couldn't find any flight for this search",
        life: 5000,
      })

      return
    }

    flights.value = response.data.comparison
    cheapest.value = response.data.cheapest
    fastest.value = response.data.fastest
  } catch (error: any) {
    loading.value = false

    if (error.response) {
      toast.add({
        severity: 'error',
        summary: 'Failed to load flights',
        detail: error.response.data,
        life: 5000,
      })

      return
    }

    toast.add({
      severity: 'error',
      summary: 'Failed to load flights',
      detail: 'Could search flight',
      life: 5000,
    })
    console.log(error)
  }
}
</script>

<template>
  <div class="flex flex-column gap-3">
    <InputGroup>
      <InputText v-model="origin" placeholder="Origin ###"></InputText>
      <InputText v-model="destination" placeholder="Destination ###"></InputText>
      <DatePicker v-model="date" dateFormat="yy-mm-dd" placeholder="YYY-MM-DD"></DatePicker>
      <Button label="Search" @click="onSearch" />
    </InputGroup>

    <div v-if="flights" class="flex">
      <Card class="w-full">
        <template #title><strong class="green">Cheapest</strong></template>
        <template #content>
          <div class="flex flex-column gap-2">
            <div>
              <span class="mr-6"><strong>Price</strong></span>
              <span>{{ cheapest.price }}</span>
            </div>
            <div>
              <span class="mr-4"><strong>Duration</strong></span>
              <span>{{ cheapest.duration }}</span>
            </div>
            <div>
              <span class="mr-4"><strong>Provider</strong></span>
              <span>{{ cheapest.agent }}</span>
            </div>
          </div>
        </template>
      </Card>

      <Divider layout="vertical" />

      <Card class="w-full">
        <template #title><strong class="green">Fastest</strong></template>
        <template #content>
          <div class="flex flex-column gap-2">
            <div>
              <span class="mr-6"><strong>Price</strong></span>
              <span>{{ fastest.price }}</span>
            </div>
            <div>
              <span class="mr-4"><strong>Duration</strong></span>
              <span>{{ fastest.duration }}</span>
            </div>
            <div>
              <span class="mr-4"><strong>Provider</strong></span>
              <span>{{ fastest.agent }}</span>
            </div>
          </div>
        </template>
      </Card>
    </div>

    <ProgressBar v-if="loading" mode="indeterminate" style="height: 6px"></ProgressBar>
    <Card class="w-full">
      <template #title><h2 class="green">Price Comparison</h2></template>
      <template #content>
        <DataTable :value="flights" tableStyle="min-width: 50rem">
          <Column field="price" header="Price"></Column>
          <Column field="duration" header="Duration"></Column>
          <Column field="agent" header="Provider"></Column>
        </DataTable>
      </template>
    </Card>
  </div>
</template>
