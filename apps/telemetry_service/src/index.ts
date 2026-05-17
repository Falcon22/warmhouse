import express, { Request, Response } from 'express';

interface Measurement {
  device_id: string;
  metric: string;
  value: number;
  unit: string;
  timestamp: string;
}

class TelemetryStore {
  private measurements: Measurement[] = [];

  add(m: Measurement): void {
    this.measurements.push(m);
  }

  query(deviceId?: string, metric?: string): Measurement[] {
    return this.measurements.filter(
      (m) =>
        (!deviceId || m.device_id === deviceId) &&
        (!metric || m.metric === metric),
    );
  }

  count(): number {
    return this.measurements.length;
  }
}

const store = new TelemetryStore();
const app = express();
app.use(express.json());

app.get('/health', (_req: Request, res: Response) => {
  res.json({ status: 'ok', stored: store.count() });
});

app.post('/telemetry', (req: Request, res: Response) => {
  const deviceId = req.body.device_id;
  const value = Number(req.body.value);

  if (!deviceId || Number.isNaN(value)) {
    return res.status(400).json({ error: 'device_id and value are required' });
  }

  const m: Measurement = {
    device_id: String(deviceId),
    metric: String(req.body.metric ?? 'temperature'),
    value,
    unit: String(req.body.unit ?? '°C'),
    timestamp: req.body.timestamp ?? new Date().toISOString(),
  };

  store.add(m);
  console.log(`telemetry stored: device=${m.device_id} ${m.metric}=${m.value}${m.unit}`);
  res.status(201).json(m);
});

app.get('/telemetry', (req: Request, res: Response) => {
  const deviceId = req.query.device_id as string | undefined;
  const metric = req.query.metric as string | undefined;
  res.json(store.query(deviceId, metric));
});

const port = Number(process.env.PORT ?? 8083);
app.listen(port, () => {
  console.log(`Telemetry service listening on :${port}`);
});
