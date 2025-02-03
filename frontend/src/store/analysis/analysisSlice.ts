import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import { analysisApi } from '@/services/api'

export interface AnalysisState {
  technicalIndicators: any[]
  marketAnalysis: any
  llmAnalysis: string
  loading: boolean
  error: string | null
}

const initialState: AnalysisState = {
  technicalIndicators: [],
  marketAnalysis: null,
  llmAnalysis: '',
  loading: false,
  error: null,
}

export const analyzeMarket = createAsyncThunk(
  'analysis/analyzeMarket',
  async (params: any) => {
    const [indicatorsResponse, analysisResponse] = await Promise.all([
      analysisApi.getIndicators(params.symbol, params),
      analysisApi.getAnalysis(params.symbol),
    ])
    return { 
      indicators: indicatorsResponse.data,
      analysis: analysisResponse.data 
    }
  }
)

export const generateReport = createAsyncThunk(
  'analysis/generateReport',
  async (params: any) => {
    const response = await analysisApi.getLLMAnalysis(params)
    return response.data
  }
)

const analysisSlice = createSlice({
  name: 'analysis',
  initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder
      .addCase(analyzeMarket.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(analyzeMarket.fulfilled, (state, action) => {
        state.loading = false
        state.technicalIndicators = action.payload.indicators
        state.marketAnalysis = action.payload.analysis
      })
      .addCase(analyzeMarket.rejected, (state, action) => {
        state.loading = false
        state.error = action.error.message || 'Analysis failed'
      })
      .addCase(generateReport.fulfilled, (state, action) => {
        state.llmAnalysis = action.payload.analysis
      })
  },
})

export default analysisSlice.reducer    