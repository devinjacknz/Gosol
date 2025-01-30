'use client';

import { useState } from 'react';
import {
  Box,
  Typography,
  TextField,
  Button,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Grid,
  Alert,
} from '@mui/material';
import {
  Add as AddIcon,
  Delete as DeleteIcon,
  Edit as EditIcon,
} from '@mui/icons-material';
import { useAnalysis } from '@/contexts/AnalysisContext';
import { INDICATOR_TYPES } from '@/config/analysis';

interface ConditionFormData {
  type: string;
  operator: string;
  value: string;
  target?: string;
}

export default function StrategyBuilder() {
  const { state, createStrategy, updateStrategy } = useAnalysis();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [openConditionDialog, setOpenConditionDialog] = useState(false);
  const [selectedCondition, setSelectedCondition] = useState<number | null>(null);
  const [conditions, setConditions] = useState<ConditionFormData[]>([]);
  const [error, setError] = useState('');

  const handleAddCondition = () => {
    setSelectedCondition(null);
    setOpenConditionDialog(true);
  };

  const handleEditCondition = (index: number) => {
    setSelectedCondition(index);
    setOpenConditionDialog(true);
  };

  const handleDeleteCondition = (index: number) => {
    setConditions(conditions.filter((_, i) => i !== index));
  };

  const handleSaveCondition = (condition: ConditionFormData) => {
    if (selectedCondition !== null) {
      const newConditions = [...conditions];
      newConditions[selectedCondition] = condition;
      setConditions(newConditions);
    } else {
      setConditions([...conditions, condition]);
    }
    setOpenConditionDialog(false);
  };

  const handleSaveStrategy = async () => {
    try {
      setError('');
      if (!name) {
        setError('Strategy name is required');
        return;
      }

      if (conditions.length === 0) {
        setError('At least one condition is required');
        return;
      }

      const strategy = {
        name,
        description,
        status: 'active',
        config: {
          timeframe: '1h',
          conditions,
          indicators: state.indicators,
          riskManagement: {
            stopLoss: 2,
            takeProfit: 4,
            trailingStop: false,
            maxDrawdown: 10,
            positionSize: 1,
          },
        },
      };

      await createStrategy(strategy);

      // Reset form
      setName('');
      setDescription('');
      setConditions([]);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save strategy');
    }
  };

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Strategy Builder
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      <Grid container spacing={2}>
        <Grid item xs={12}>
          <TextField
            fullWidth
            label="Strategy Name"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
        </Grid>

        <Grid item xs={12}>
          <TextField
            fullWidth
            multiline
            rows={3}
            label="Description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </Grid>

        <Grid item xs={12}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
            <Typography variant="subtitle1">Conditions</Typography>
            <Button
              startIcon={<AddIcon />}
              onClick={handleAddCondition}
              size="small"
            >
              Add Condition
            </Button>
          </Box>

          <List>
            {conditions.map((condition, index) => (
              <ListItem
                key={index}
                sx={{ bgcolor: 'background.paper', mb: 1, borderRadius: 1 }}
              >
                <ListItemText
                  primary={`${condition.type} ${condition.operator} ${condition.value}`}
                  secondary={condition.target}
                />
                <ListItemSecondaryAction>
                  <IconButton
                    edge="end"
                    aria-label="edit"
                    onClick={() => handleEditCondition(index)}
                  >
                    <EditIcon />
                  </IconButton>
                  <IconButton
                    edge="end"
                    aria-label="delete"
                    onClick={() => handleDeleteCondition(index)}
                  >
                    <DeleteIcon />
                  </IconButton>
                </ListItemSecondaryAction>
              </ListItem>
            ))}
          </List>
        </Grid>

        <Grid item xs={12}>
          <Button
            fullWidth
            variant="contained"
            color="primary"
            onClick={handleSaveStrategy}
          >
            Save Strategy
          </Button>
        </Grid>
      </Grid>

      <ConditionDialog
        open={openConditionDialog}
        onClose={() => setOpenConditionDialog(false)}
        onSave={handleSaveCondition}
        initialData={
          selectedCondition !== null ? conditions[selectedCondition] : undefined
        }
      />
    </Box>
  );
}

interface ConditionDialogProps {
  open: boolean;
  onClose: () => void;
  onSave: (condition: ConditionFormData) => void;
  initialData?: ConditionFormData;
}

function ConditionDialog({
  open,
  onClose,
  onSave,
  initialData,
}: ConditionDialogProps) {
  const [type, setType] = useState(initialData?.type || 'PRICE');
  const [operator, setOperator] = useState(initialData?.operator || 'GREATER');
  const [value, setValue] = useState(initialData?.value || '');
  const [target, setTarget] = useState(initialData?.target || '');

  const handleSave = () => {
    onSave({
      type,
      operator,
      value,
      target: type === 'INDICATOR' ? target : undefined,
    });
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>
        {initialData ? 'Edit Condition' : 'Add Condition'}
      </DialogTitle>
      <DialogContent>
        <Grid container spacing={2} sx={{ mt: 1 }}>
          <Grid item xs={12}>
            <FormControl fullWidth>
              <InputLabel>Type</InputLabel>
              <Select
                value={type}
                label="Type"
                onChange={(e) => setType(e.target.value)}
              >
                <MenuItem value="PRICE">Price</MenuItem>
                <MenuItem value="INDICATOR">Indicator</MenuItem>
                <MenuItem value="VOLUME">Volume</MenuItem>
                <MenuItem value="TIME">Time</MenuItem>
              </Select>
            </FormControl>
          </Grid>

          {type === 'INDICATOR' && (
            <Grid item xs={12}>
              <FormControl fullWidth>
                <InputLabel>Indicator</InputLabel>
                <Select
                  value={target}
                  label="Indicator"
                  onChange={(e) => setTarget(e.target.value)}
                >
                  {Object.entries(INDICATOR_TYPES).map(([key, value]) => (
                    <MenuItem key={value} value={value}>
                      {key}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
          )}

          <Grid item xs={12}>
            <FormControl fullWidth>
              <InputLabel>Operator</InputLabel>
              <Select
                value={operator}
                label="Operator"
                onChange={(e) => setOperator(e.target.value)}
              >
                <MenuItem value="GREATER">Greater Than</MenuItem>
                <MenuItem value="LESS">Less Than</MenuItem>
                <MenuItem value="EQUAL">Equal To</MenuItem>
                <MenuItem value="CROSS_ABOVE">Crosses Above</MenuItem>
                <MenuItem value="CROSS_BELOW">Crosses Below</MenuItem>
              </Select>
            </FormControl>
          </Grid>

          <Grid item xs={12}>
            <TextField
              fullWidth
              label="Value"
              value={value}
              onChange={(e) => setValue(e.target.value)}
            />
          </Grid>
        </Grid>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleSave} variant="contained">
          Save
        </Button>
      </DialogActions>
    </Dialog>
  );
} 