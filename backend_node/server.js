const http = require('http');
const express = require('express');
const cors = require('cors');
const bcrypt = require('bcryptjs');
const jwt = require('jsonwebtoken');
const { createClient: createSupabaseClient } = require('@supabase/supabase-js');
const { WebSocketServer } = require('ws');
const { createClient } = require('redis');
const { Pool } = require('pg');
const { randomUUID } = require('crypto');

const app = express();
const server = http.createServer(app);

const PORT = Number(process.env.PORT || 8080);
const DATABASE_URL = process.env.DATABASE_URL || '';
const ML_SERVICE_URL = (process.env.ML_SERVICE_URL || 'http://localhost:8000').replace(/\/$/, '');
const REDIS_URL = process.env.REDIS_URL || '';
const JWT_SECRET = process.env.JWT_SECRET || 'dev-secret';
const SUPABASE_URL = process.env.SUPABASE_URL || '';
const SUPABASE_ANON_KEY = process.env.SUPABASE_ANON_KEY || '';
const SUPABASE_SERVICE_ROLE_KEY = process.env.SUPABASE_SERVICE_ROLE_KEY || '';
const REALTIME_CHANNEL = 'smartrestro:queue-events';

const ACTIVE_STATUSES = new Set(['WAITING', 'CONFIRMED']);
const VALID_TRANSITIONS = {
  WAITING: new Set(['CONFIRMED', 'CANCELLED']),
  CONFIRMED: new Set(['ARRIVED', 'EXPIRED', 'CANCELLED']),
  ARRIVED: new Set(['SEATED', 'CANCELLED']),
};

app.use(cors());
app.use(express.json());

let pool = null;
if (DATABASE_URL) {
  pool = new Pool({
    connectionString: DATABASE_URL,
    max: 15,
    idleTimeoutMillis: 30000,
  });
}

const supabaseEnabled = Boolean(SUPABASE_URL && SUPABASE_ANON_KEY);
const supabaseAuthClient = supabaseEnabled
  ? createSupabaseClient(SUPABASE_URL, SUPABASE_ANON_KEY, {
      auth: { persistSession: false, autoRefreshToken: false },
    })
  : null;
const supabaseAdminClient = SUPABASE_URL && SUPABASE_SERVICE_ROLE_KEY
  ? createSupabaseClient(SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY, {
      auth: { persistSession: false, autoRefreshToken: false },
    })
  : null;

const pubClient = REDIS_URL ? createClient({ url: REDIS_URL }) : null;
const subClient = REDIS_URL ? createClient({ url: REDIS_URL }) : null;
const sseClients = new Map();
const wsSubscriptions = new Map();
const realtimeStatus = {
  redisConnected: false,
};

const seedRestaurant = {
  id: '11111111-1111-1111-1111-111111111111',
  name: 'SmartRestro Flagship',
  address: 'MG Road, Pune',
  latitude: 18.5204,
  longitude: 73.8567,
  capacity: 48,
  current_occupancy: 31,
  average_meal_minutes: 45,
  created_at: new Date().toISOString(),
};

const seedRestaurants = [
  seedRestaurant,
  {
    id: '44444444-4444-4444-8444-444444444444',
    name: 'Lakeview Tandoor',
    address: 'JM Road, Pune',
    latitude: 18.5272,
    longitude: 73.8424,
    capacity: 62,
    current_occupancy: 37,
    average_meal_minutes: 52,
    created_at: new Date().toISOString(),
  },
  {
    id: '55555555-5555-4555-8555-555555555555',
    name: 'Garden Terrace Cafe',
    address: 'FC Road, Pune',
    latitude: 18.5236,
    longitude: 73.8418,
    capacity: 40,
    current_occupancy: 19,
    average_meal_minutes: 38,
    created_at: new Date().toISOString(),
  },
];

const adminUser = {
  id: 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa',
  name: 'Queue Admin',
  email: 'admin@smartrestro.local',
  role: 'restaurant_admin',
  password_hash: bcrypt.hashSync('admin123', 10),
  created_at: new Date().toISOString(),
};

const memory = {
  users: [
    adminUser,
    {
      id: randomUUID(),
      name: 'Demo Customer',
      email: 'demo_customer@example.com',
      role: 'customer',
      password_hash: bcrypt.hashSync('customer123', 10),
      created_at: new Date().toISOString(),
    },
    {
      id: randomUUID(),
      name: 'Demo Staff',
      email: 'demo_staff@example.com',
      role: 'restaurant_staff',
      password_hash: bcrypt.hashSync('staff123', 10),
      created_at: new Date().toISOString(),
    },
    {
      id: randomUUID(),
      name: 'Demo Admin',
      email: 'demo_admin@example.com',
      role: 'restaurant_admin',
      password_hash: bcrypt.hashSync('admin123', 10),
      created_at: new Date().toISOString(),
    }
  ],
  restaurants: seedRestaurants,
  queueEntries: [
    {
      id: '22222222-2222-4222-8222-222222222222',
      restaurant_id: seedRestaurant.id,
      guest_name: 'John Doe',
      party_size: 4,
      status: 'WAITING',
      phone_number: '',
      estimated_wait_time: 18,
      position: 1,
      source_channel: 'dashboard',
      joined_at: new Date(Date.now() - 18 * 60000).toISOString(),
      updated_at: new Date(Date.now() - 18 * 60000).toISOString(),
    },
    {
      id: '33333333-3333-4333-8333-333333333333',
      restaurant_id: seedRestaurant.id,
      guest_name: 'Alice Smith',
      party_size: 2,
      status: 'CONFIRMED',
      phone_number: '',
      estimated_wait_time: 9,
      position: 2,
      source_channel: 'mobile',
      joined_at: new Date(Date.now() - 9 * 60000).toISOString(),
      updated_at: new Date(Date.now() - 4 * 60000).toISOString(),
    },
  ],
  queueEvents: {
    '22222222-2222-4222-8222-222222222222': [
      {
        id: randomUUID(),
        queue_entry_id: '22222222-2222-4222-8222-222222222222',
        restaurant_id: seedRestaurant.id,
        event_type: 'queue.joined',
        from_status: null,
        to_status: 'WAITING',
        actor_user_id: '',
        created_at: new Date(Date.now() - 18 * 60000).toISOString(),
      },
    ],
    '33333333-3333-4333-8333-333333333333': [
      {
        id: randomUUID(),
        queue_entry_id: '33333333-3333-4333-8333-333333333333',
        restaurant_id: seedRestaurant.id,
        event_type: 'queue.joined',
        from_status: null,
        to_status: 'WAITING',
        actor_user_id: '',
        created_at: new Date(Date.now() - 9 * 60000).toISOString(),
      },
      {
        id: randomUUID(),
        queue_entry_id: '33333333-3333-4333-8333-333333333333',
        restaurant_id: seedRestaurant.id,
        event_type: 'queue.state_changed',
        from_status: 'WAITING',
        to_status: 'CONFIRMED',
        actor_user_id: '',
        created_at: new Date(Date.now() - 4 * 60000).toISOString(),
      },
    ],
  },
};

bootRedis().catch((error) => {
  console.error('Redis boot failed:', error.message);
});

app.get('/health', async (_req, res) => {
  const status = {
    status: 'healthy',
    backend: 'node',
    storage: pool ? 'postgres' : 'memory',
    redis_configured: Boolean(REDIS_URL),
    redis_connected: realtimeStatus.redisConnected,
    supabase_configured: supabaseEnabled,
  };

  if (!pool) {
    res.json(status);
    return;
  }

  try {
    await pool.query('SELECT 1');
    res.json(status);
  } catch (_error) {
    res.status(503).json({ ...status, status: 'degraded', error: 'database unavailable' });
  }
});

app.post('/api/v1/auth/signup', async (req, res) => {
  const { name, email, password, role = 'customer' } = req.body || {};
  if (!name || !email || !password) {
    res.status(400).json({ error: 'name, email, and password are required' });
    return;
  }

  try {
    if (supabaseAuthClient) {
      const { data, error } = await supabaseAuthClient.auth.signUp({
        email: email.toLowerCase(),
        password,
        options: {
          data: { name, role },
        },
      });

      if (error) {
        res.status(400).json({ error: error.message });
        return;
      }

      if (data.session) {
        res.status(201).json({
          data: {
            token: data.session.access_token,
            user: normalizeSupabaseUser(data.user),
            provider: 'supabase',
          },
        });
        return;
      }

      res.status(202).json({
        data: {
          token: '',
          user: normalizeSupabaseUser(data.user),
          provider: 'supabase',
          requires_confirmation: true,
          message: 'Sign up created. Confirm the email in Supabase before logging in.',
        },
      });
      return;
    }

    const existing = await findUserByEmail(email);
    if (existing) {
      res.status(409).json({ error: 'email already registered' });
      return;
    }

    const user = await createUser({
      id: randomUUID(),
      name,
      email: email.toLowerCase(),
      password_hash: await bcrypt.hash(password, 10),
      role,
      created_at: new Date().toISOString(),
    });

    res.status(201).json(buildAuthResponse(user));
  } catch (error) {
    handleError(res, error);
  }
});

app.post('/api/v1/auth/login', async (req, res) => {
  const { email, password } = req.body || {};
  if (!email || !password) {
    res.status(400).json({ error: 'email and password are required' });
    return;
  }

  try {
    if (supabaseAuthClient) {
      const { data, error } = await supabaseAuthClient.auth.signInWithPassword({
        email: email.toLowerCase(),
        password,
      });

      if (error || !data.session || !data.user) {
        res.status(401).json({ error: error?.message || 'invalid credentials' });
        return;
      }

      res.json({
        data: {
          token: data.session.access_token,
          user: normalizeSupabaseUser(data.user),
          provider: 'supabase',
        },
      });
      return;
    }

    const user = await findUserByEmail(email);
    if (!user) {
      res.status(401).json({ error: 'invalid credentials' });
      return;
    }

    const valid = await bcrypt.compare(password, user.password_hash);
    if (!valid) {
      res.status(401).json({ error: 'invalid credentials' });
      return;
    }

    res.json(buildAuthResponse(user));
  } catch (error) {
    handleError(res, error);
  }
});

app.get('/api/v1/auth/me', authenticate, async (req, res) => {
  res.json({ data: sanitizeUser(req.user) });
});

app.get('/api/v1/realtime/sse', authenticate, (req, res) => {
  const restaurantId = String(req.query.restaurant_id || '');
  if (!restaurantId) {
    res.status(400).json({ error: 'restaurant_id is required' });
    return;
  }

  res.writeHead(200, {
    'Content-Type': 'text/event-stream',
    'Cache-Control': 'no-cache',
    Connection: 'keep-alive',
  });
  res.write(`data: ${JSON.stringify({ type: 'connected', restaurant_id: restaurantId })}\n\n`);

  const bucket = sseClients.get(restaurantId) || new Set();
  bucket.add(res);
  sseClients.set(restaurantId, bucket);

  req.on('close', () => {
    const clients = sseClients.get(restaurantId);
    if (!clients) {
      return;
    }
    clients.delete(res);
    if (clients.size === 0) {
      sseClients.delete(restaurantId);
    }
  });
});

app.use('/api/v1', authenticate);

app.get('/api/v1/restaurants', async (_req, res) => {
  try {
    res.json({ data: await listRestaurants() });
  } catch (error) {
    handleError(res, error);
  }
});

app.get('/api/v1/restaurants/nearby', async (req, res) => {
  const lat = Number(req.query.lat);
  const lng = Number(req.query.lng);
  const radiusKm = Number(req.query.radius || 2);

  if (Number.isNaN(lat) || Number.isNaN(lng) || Number.isNaN(radiusKm)) {
    res.status(400).json({ error: 'invalid lat, lng, or radius' });
    return;
  }

  try {
    res.json({ data: await listNearbyRestaurants(lat, lng, radiusKm) });
  } catch (error) {
    handleError(res, error);
  }
});

app.get('/api/v1/restaurants/:id', async (req, res) => {
  try {
    const restaurant = await getRestaurant(req.params.id);
    if (!restaurant) {
      res.status(404).json({ error: 'restaurant not found' });
      return;
    }
    res.json({ data: restaurant });
  } catch (error) {
    handleError(res, error);
  }
});

app.get('/api/v1/restaurants/:id/queue', async (req, res) => {
  try {
    const restaurant = await getRestaurant(req.params.id);
    if (!restaurant) {
      res.status(404).json({ error: 'restaurant not found' });
      return;
    }
    const queue = await getQueueByRestaurant(req.params.id);
    res.json({ data: recalculateQueue(queue) });
  } catch (error) {
    handleError(res, error);
  }
});

app.get('/api/v1/restaurants/:id/queue-summary', async (req, res) => {
  try {
    const summary = await getQueueSummary(req.params.id);
    if (!summary) {
      res.status(404).json({ error: 'restaurant not found' });
      return;
    }
    res.json({ data: summary });
  } catch (error) {
    handleError(res, error);
  }
});

app.post('/api/v1/restaurants/:id/queue/join', async (req, res) => {
  const { guest_name, party_size, phone_number = '', source_channel = 'dashboard' } = req.body || {};

  if (!guest_name || !party_size || Number(party_size) < 1) {
    res.status(400).json({ error: 'guest_name and party_size are required' });
    return;
  }

  try {
    const restaurant = await getRestaurant(req.params.id);
    if (!restaurant) {
      res.status(404).json({ error: 'restaurant not found' });
      return;
    }

    const queue = recalculateQueue(await getQueueByRestaurant(req.params.id));
    const prediction = await predictWaitTime({
      restaurant_id: restaurant.id,
      queue_depth: queue.filter((entry) => ACTIVE_STATUSES.has(entry.status)).length + 1,
      party_size: Number(party_size),
      average_meal_minutes: restaurant.average_meal_minutes,
      occupancy: restaurant.current_occupancy,
      capacity: restaurant.capacity,
      seating_velocity: 1.0,
      historical_no_show_rate: 0.08,
    });

    const now = new Date().toISOString();
    const entry = {
      id: randomUUID(),
      restaurant_id: restaurant.id,
      guest_name,
      party_size: Number(party_size),
      status: 'WAITING',
      phone_number,
      estimated_wait_time: prediction.predicted_wait_time,
      position: queue.filter((item) => ACTIVE_STATUSES.has(item.status)).length + 1,
      source_channel,
      joined_at: now,
      updated_at: now,
    };

    const event = {
      id: randomUUID(),
      queue_entry_id: entry.id,
      restaurant_id: restaurant.id,
      event_type: 'queue.joined',
      from_status: null,
      to_status: 'WAITING',
      actor_user_id: req.user.id,
      created_at: now,
    };

    await createQueueEntry(entry, event);
    await broadcastQueueEvent({
      type: 'queue.joined',
      restaurant_id: restaurant.id,
      entry,
      summary: await getQueueSummary(restaurant.id),
    });

    res.status(201).json({ data: entry });
  } catch (error) {
    handleError(res, error);
  }
});

app.get('/api/v1/queue/:id', async (req, res) => {
  try {
    const entry = await getQueueEntry(req.params.id);
    if (!entry) {
      res.status(404).json({ error: 'queue entry not found' });
      return;
    }
    res.json({ data: { entry: await refreshEntry(entry), events: await getQueueEvents(req.params.id) } });
  } catch (error) {
    handleError(res, error);
  }
});

app.patch('/api/v1/queue/:id/status', async (req, res) => {
  const nextStatus = req.body?.status;
  if (!nextStatus) {
    res.status(400).json({ error: 'status is required' });
    return;
  }

  try {
    const entry = await getQueueEntry(req.params.id);
    if (!entry) {
      res.status(404).json({ error: 'queue entry not found' });
      return;
    }

    if (entry.status !== nextStatus) {
      const allowed = VALID_TRANSITIONS[entry.status] && VALID_TRANSITIONS[entry.status].has(nextStatus);
      if (!allowed) {
        res.status(400).json({ error: `invalid status transition from ${entry.status} to ${nextStatus}` });
        return;
      }
    }

    const updated = {
      ...entry,
      status: nextStatus,
      updated_at: new Date().toISOString(),
    };

    const event = {
      id: randomUUID(),
      queue_entry_id: updated.id,
      restaurant_id: updated.restaurant_id,
      event_type: 'queue.state_changed',
      from_status: entry.status,
      to_status: nextStatus,
      actor_user_id: req.user.id,
      created_at: updated.updated_at,
    };

    await updateQueueEntry(updated, event);
    const refreshed = await refreshEntry(updated);
    await broadcastQueueEvent({
      type: 'queue.state_changed',
      restaurant_id: updated.restaurant_id,
      entry: refreshed,
      summary: await getQueueSummary(updated.restaurant_id),
    });

    res.json({ data: refreshed });
  } catch (error) {
    handleError(res, error);
  }
});

app.get('/api/v1/restaurants/recommend', async (req, res) => {
  const lat = Number(req.query.lat);
  const lng = Number(req.query.lng);
  const radiusKm = Number(req.query.radius || 3);
  const partySize = Number(req.query.party_size || 2);

  if (Number.isNaN(lat) || Number.isNaN(lng) || Number.isNaN(radiusKm) || Number.isNaN(partySize)) {
    res.status(400).json({ error: 'invalid lat, lng, radius, or party_size' });
    return;
  }

  try {
    const restaurants = await listNearbyRestaurants(lat, lng, radiusKm);
    const candidateRestaurants = await Promise.all(
      restaurants.map(async (restaurant) => {
        const summary = await getQueueSummary(restaurant.id);
        const waitPrediction = await predictWaitTime({
          restaurant_id: restaurant.id,
          queue_depth: summary ? summary.active_queue_count + 1 : 1,
          party_size: partySize,
          average_meal_minutes: restaurant.average_meal_minutes,
          occupancy: restaurant.current_occupancy,
          capacity: restaurant.capacity,
          seating_velocity: 1.0,
          historical_no_show_rate: 0.08,
        });

        return {
          ...restaurant,
          queue_summary: summary,
          predicted_wait_time: waitPrediction.predicted_wait_time,
          wait_confidence: waitPrediction.confidence,
        };
      })
    );

    res.json({
      data: await recommendRestaurants({
        latitude: lat,
        longitude: lng,
        radius_km: radiusKm,
        party_size: partySize,
        candidate_restaurants: candidateRestaurants,
      }),
    });
  } catch (error) {
    handleError(res, error);
  }
});

async function authenticate(req, res, next) {
  const header = req.headers.authorization || '';
  const token = header.startsWith('Bearer ') ? header.slice(7) : String(req.query.token || '');
  if (!token) {
    res.status(401).json({ error: 'authorization token required' });
    return;
  }

  try {
    const payload = jwt.verify(token, JWT_SECRET);
    req.user = payload;
    next();
    return;
  } catch (_error) {
    if (supabaseAuthClient) {
      const { data, error } = await supabaseAuthClient.auth.getUser(token);
      if (!error && data.user) {
        req.user = normalizeSupabaseUser(data.user);
        next();
        return;
      }
    }

    res.status(401).json({ error: 'invalid token' });
  }
}

function buildAuthResponse(user) {
  const safeUser = sanitizeUser(user);
  const token = jwt.sign(safeUser, JWT_SECRET, { expiresIn: '12h' });
  return { data: { token, user: safeUser } };
}

function sanitizeUser(user) {
  return {
    id: user.id,
    name: user.name,
    email: user.email,
    role: user.role || 'customer',
  };
}

function normalizeSupabaseUser(user) {
  return {
    id: user?.id || '',
    name: user?.user_metadata?.name || user?.email || 'Supabase User',
    email: user?.email || '',
    role: user?.user_metadata?.role || 'customer',
  };
}

async function findUserByEmail(email) {
  const normalized = String(email).toLowerCase();
  if (!pool) {
    return memory.users.find((user) => user.email === normalized) || null;
  }

  const result = await pool.query(`
    SELECT id::text, name, email, password_hash, role, created_at
    FROM app_users
    WHERE email = $1
  `, [normalized]);

  return result.rows[0] || null;
}

async function createUser(user) {
  if (!pool) {
    memory.users.push(user);
    return user;
  }

  await pool.query(`
    INSERT INTO app_users (id, name, email, password_hash, role, created_at, updated_at)
    VALUES ($1::uuid, $2, $3, $4, $5, $6, $6)
  `, [user.id, user.name, user.email, user.password_hash, user.role, user.created_at]);

  return user;
}

async function listRestaurants() {
  if (!pool) {
    return memory.restaurants;
  }

  const result = await pool.query(`
    SELECT id::text, name, COALESCE(address, '') AS address,
           COALESCE(ST_Y(location::geometry), 0) AS latitude,
           COALESCE(ST_X(location::geometry), 0) AS longitude,
           capacity, current_occupancy, average_meal_duration AS average_meal_minutes,
           created_at
    FROM restaurants
    ORDER BY created_at ASC
  `);

  return result.rows;
}

async function listNearbyRestaurants(lat, lng, radiusKm) {
  if (!pool) {
    return memory.restaurants
      .map((restaurant) => ({
        ...restaurant,
        distance_km: round2(haversineKm(lat, lng, restaurant.latitude, restaurant.longitude)),
      }))
      .filter((restaurant) => restaurant.distance_km <= radiusKm)
      .sort((a, b) => a.distance_km - b.distance_km);
  }

  const result = await pool.query(`
    SELECT id::text, name, COALESCE(address, '') AS address,
           COALESCE(ST_Y(location::geometry), 0) AS latitude,
           COALESCE(ST_X(location::geometry), 0) AS longitude,
           capacity, current_occupancy, average_meal_duration AS average_meal_minutes,
           created_at,
           ST_Distance(
             location,
             ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography
           ) / 1000.0 AS distance_km
    FROM restaurants
    WHERE location IS NOT NULL
      AND ST_DWithin(
        location,
        ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography,
        $3 * 1000.0
      )
    ORDER BY distance_km ASC
  `, [lat, lng, radiusKm]);

  return result.rows.map((row) => ({ ...row, distance_km: Number(row.distance_km) }));
}

async function getRestaurant(id) {
  if (!pool) {
    return memory.restaurants.find((restaurant) => restaurant.id === id) || null;
  }

  const result = await pool.query(`
    SELECT id::text, name, COALESCE(address, '') AS address,
           COALESCE(ST_Y(location::geometry), 0) AS latitude,
           COALESCE(ST_X(location::geometry), 0) AS longitude,
           capacity, current_occupancy, average_meal_duration AS average_meal_minutes,
           created_at
    FROM restaurants
    WHERE id = $1::uuid
  `, [id]);

  return result.rows[0] || null;
}

async function getQueueByRestaurant(restaurantId) {
  if (!pool) {
    return memory.queueEntries.filter((entry) => entry.restaurant_id === restaurantId);
  }

  const result = await pool.query(`
    SELECT id::text, restaurant_id::text, guest_name, party_size,
           status::text, COALESCE(phone_number, '') AS phone_number,
           estimated_wait_time, position, source_channel, joined_at, updated_at
    FROM queue_entries
    WHERE restaurant_id = $1::uuid
    ORDER BY joined_at ASC
  `, [restaurantId]);

  return result.rows;
}

async function getQueueEntry(id) {
  if (!pool) {
    return memory.queueEntries.find((entry) => entry.id === id) || null;
  }

  const result = await pool.query(`
    SELECT id::text, restaurant_id::text, guest_name, party_size,
           status::text, COALESCE(phone_number, '') AS phone_number,
           estimated_wait_time, position, source_channel, joined_at, updated_at
    FROM queue_entries
    WHERE id = $1::uuid
  `, [id]);

  return result.rows[0] || null;
}

async function getQueueEvents(queueEntryId) {
  if (!pool) {
    return memory.queueEvents[queueEntryId] || [];
  }

  const result = await pool.query(`
    SELECT id::text, queue_entry_id::text, restaurant_id::text,
           event_type, from_status::text, to_status::text,
           COALESCE(actor_user_id::text, '') AS actor_user_id, created_at
    FROM queue_events
    WHERE queue_entry_id = $1::uuid
    ORDER BY created_at ASC
  `, [queueEntryId]);

  return result.rows;
}

async function createQueueEntry(entry, event) {
  if (!pool) {
    memory.queueEntries.push(entry);
    memory.queueEvents[entry.id] = [...(memory.queueEvents[entry.id] || []), event];
    return entry;
  }

  const client = await pool.connect();
  try {
    await client.query('BEGIN');
    await client.query(`
      INSERT INTO queue_entries (
        id, restaurant_id, guest_name, party_size, status,
        phone_number, estimated_wait_time, position, source_channel, joined_at, updated_at
      ) VALUES (
        $1::uuid, $2::uuid, $3, $4, $5::queue_status,
        NULLIF($6, ''), $7, $8, $9, $10, $11
      )
    `, [
      entry.id,
      entry.restaurant_id,
      entry.guest_name,
      entry.party_size,
      entry.status,
      entry.phone_number,
      entry.estimated_wait_time,
      entry.position,
      entry.source_channel,
      entry.joined_at,
      entry.updated_at,
    ]);
    await insertQueueEvent(client, event);
    await client.query('COMMIT');
    return entry;
  } catch (error) {
    await client.query('ROLLBACK');
    throw error;
  } finally {
    client.release();
  }
}

async function updateQueueEntry(entry, event) {
  if (!pool) {
    const index = memory.queueEntries.findIndex((item) => item.id === entry.id);
    if (index >= 0) {
      memory.queueEntries[index] = entry;
    }
    memory.queueEvents[entry.id] = [...(memory.queueEvents[entry.id] || []), event];
    return entry;
  }

  const client = await pool.connect();
  try {
    await client.query('BEGIN');
    await client.query(`
      UPDATE queue_entries
      SET status = $2::queue_status,
          estimated_wait_time = $3,
          position = $4,
          updated_at = $5
      WHERE id = $1::uuid
    `, [
      entry.id,
      entry.status,
      entry.estimated_wait_time,
      entry.position || 0,
      entry.updated_at,
    ]);
    await insertQueueEvent(client, event);
    await client.query('COMMIT');
    return entry;
  } catch (error) {
    await client.query('ROLLBACK');
    throw error;
  } finally {
    client.release();
  }
}

async function insertQueueEvent(client, event) {
  await client.query(`
    INSERT INTO queue_events (
      id, queue_entry_id, restaurant_id, event_type,
      from_status, to_status, actor_user_id, created_at
    ) VALUES (
      $1::uuid, $2::uuid, $3::uuid, $4,
      $5::queue_status, $6::queue_status, NULLIF($7, '')::uuid, $8
    )
  `, [
    event.id,
    event.queue_entry_id,
    event.restaurant_id,
    event.event_type,
    event.from_status,
    event.to_status,
    event.actor_user_id || '',
    event.created_at,
  ]);
}

async function getQueueSummary(restaurantId) {
  const restaurant = await getRestaurant(restaurantId);
  if (!restaurant) {
    return null;
  }

  const queue = recalculateQueue(await getQueueByRestaurant(restaurantId));
  const waitingCount = queue.filter((entry) => entry.status === 'WAITING').length;
  const confirmedCount = queue.filter((entry) => entry.status === 'CONFIRMED').length;
  const activeEntries = queue.filter((entry) => ACTIVE_STATUSES.has(entry.status));
  const averageWaitMinutes = activeEntries.length
    ? Math.round(activeEntries.reduce((total, entry) => total + Number(entry.estimated_wait_time || 0), 0) / activeEntries.length)
    : 0;

  return {
    restaurant_id: restaurantId,
    active_queue_count: activeEntries.length,
    waiting_count: waitingCount,
    confirmed_count: confirmedCount,
    average_wait_minutes: averageWaitMinutes,
    current_occupancy: Number(restaurant.current_occupancy || 0),
  };
}

function recalculateQueue(queueEntries) {
  const sorted = [...queueEntries].sort((a, b) => new Date(a.joined_at) - new Date(b.joined_at));
  let position = 0;
  return sorted.map((entry) => {
    if (ACTIVE_STATUSES.has(entry.status)) {
      position += 1;
      return { ...entry, position };
    }
    return { ...entry, position: 0 };
  });
}

async function refreshEntry(entry) {
  const queue = recalculateQueue(await getQueueByRestaurant(entry.restaurant_id));
  return queue.find((item) => item.id === entry.id) || entry;
}

async function predictWaitTime(payload) {
  try {
    const response = await fetch(`${ML_SERVICE_URL}/predict/wait-time`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    if (!response.ok) {
      throw new Error(`predictor status ${response.status}`);
    }
    return await response.json();
  } catch (_error) {
    return localWaitPrediction(payload);
  }
}

async function recommendRestaurants(payload) {
  try {
    const response = await fetch(`${ML_SERVICE_URL}/recommend/restaurants`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    if (!response.ok) {
      throw new Error(`recommendation status ${response.status}`);
    }
    return await response.json();
  } catch (_error) {
    return {
      recommended_restaurants: [...payload.candidate_restaurants]
        .sort((a, b) => a.predicted_wait_time - b.predicted_wait_time)
        .slice(0, 3),
      strategy: 'node-fallback',
      rationale: ['ML recommendation service unavailable; ranked by predicted wait time.'],
      timestamp: new Date().toISOString(),
    };
  }
}

function localWaitPrediction(payload) {
  const occupancyRatio = payload.capacity > 0 ? payload.occupancy / payload.capacity : 0;
  const queuePressure = payload.queue_depth * payload.average_meal_minutes * 0.35;
  const partyFactor = payload.party_size >= 5 ? 1.2 : 1.0;
  const noShowRelief = 1.0 - payload.historical_no_show_rate * 0.25;
  const velocityFactor = 1.0 / payload.seating_velocity;
  return {
    predicted_wait_time: Math.max(5, Math.round(queuePressure * (1 + occupancyRatio) * partyFactor * noShowRelief * velocityFactor)),
    confidence: 0.72,
    model_version: 'heuristic-ewma-v1',
    timestamp: new Date().toISOString(),
  };
}

async function bootRedis() {
  if (!pubClient || !subClient) {
    return;
  }

  pubClient.on('error', (error) => console.error('Redis pub error:', error.message));
  subClient.on('error', (error) => console.error('Redis sub error:', error.message));

  await pubClient.connect();
  await subClient.connect();
  realtimeStatus.redisConnected = true;

  await subClient.subscribe(REALTIME_CHANNEL, (message) => {
    try {
      fanOutRealtime(JSON.parse(message));
    } catch (error) {
      console.error('Failed to parse realtime payload:', error.message);
    }
  });
}

async function broadcastQueueEvent(payload) {
  const event = {
    ...payload,
    sent_at: new Date().toISOString(),
  };

  if (realtimeStatus.redisConnected && pubClient) {
    await pubClient.publish(REALTIME_CHANNEL, JSON.stringify(event));
    return;
  }

  fanOutRealtime(event);
}

function fanOutRealtime(payload) {
  const restaurantId = payload.restaurant_id;
  if (!restaurantId) {
    return;
  }

  const sseBucket = sseClients.get(restaurantId);
  if (sseBucket) {
    for (const client of sseBucket) {
      client.write(`data: ${JSON.stringify(payload)}\n\n`);
    }
  }

  for (const [socket, subscriptions] of wsSubscriptions.entries()) {
    if (socket.readyState !== 1) {
      wsSubscriptions.delete(socket);
      continue;
    }
    if (subscriptions.has(restaurantId)) {
      socket.send(JSON.stringify(payload));
    }
  }
}

const wss = new WebSocketServer({ server, path: '/api/v1/realtime/ws' });
wss.on('connection', (socket, req) => {
  const url = new URL(req.url, `http://${req.headers.host}`);
  const token = url.searchParams.get('token') || '';

  try {
    jwt.verify(token, JWT_SECRET);
  } catch (_error) {
    socket.close(4001, 'invalid token');
    return;
  }

  wsSubscriptions.set(socket, new Set());
  socket.send(JSON.stringify({ type: 'connected' }));

  socket.on('message', (message) => {
    try {
      const payload = JSON.parse(String(message));
      if (payload.type === 'subscribe' && payload.restaurant_id) {
        wsSubscriptions.get(socket)?.add(String(payload.restaurant_id));
        socket.send(JSON.stringify({ type: 'subscribed', restaurant_id: String(payload.restaurant_id) }));
      }
    } catch (_error) {
      socket.send(JSON.stringify({ type: 'error', message: 'invalid payload' }));
    }
  });

  socket.on('close', () => {
    wsSubscriptions.delete(socket);
  });
});

function haversineKm(lat1, lon1, lat2, lon2) {
  const earthRadius = 6371;
  const toRadians = (value) => (value * Math.PI) / 180;
  const dLat = toRadians(lat2 - lat1);
  const dLon = toRadians(lon2 - lon1);
  const left = toRadians(lat1);
  const right = toRadians(lat2);
  const a = Math.sin(dLat / 2) ** 2 + Math.sin(dLon / 2) ** 2 * Math.cos(left) * Math.cos(right);
  return earthRadius * (2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a)));
}

function round2(value) {
  return Math.round(value * 100) / 100;
}

function handleError(res, error) {
  console.error(error);
  res.status(500).json({ error: 'internal server error' });
}

server.listen(PORT, () => {
  console.log(`SmartRestro Node API listening on http://localhost:${PORT}`);
});
