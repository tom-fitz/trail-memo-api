# TrailMemo API - Deployment Guide

Complete guide for deploying TrailMemo API to production.

## Railway.app Deployment (Recommended)

Railway is a modern platform that makes deployment simple. It's perfect for Go applications and includes PostgreSQL.

### Why Railway?

- ✅ Free PostgreSQL included
- ✅ Auto-deployment on git push
- ✅ Environment variable management
- ✅ HTTPS out of the box
- ✅ Easy scaling
- ✅ Good free tier

### Step 1: Prepare Your Repository

1. **Ensure all files are committed:**

```bash
git status
git add .
git commit -m "Ready for deployment"
```

2. **Push to GitHub:**

```bash
git push origin main
```

### Step 2: Create Railway Project

1. Go to [railway.app](https://railway.app/)
2. Sign up with GitHub (free)
3. Click "New Project"
4. Select "Deploy from GitHub repo"
5. Choose your `trailmemo-api` repository
6. Railway will automatically detect Go and start building

### Step 3: Add PostgreSQL Database

1. In your Railway project, click "+ New"
2. Select "Database" → "PostgreSQL"
3. Railway automatically provisions and connects it
4. The `DATABASE_URL` environment variable is automatically set

### Step 4: Configure Environment Variables

1. Click on your service (the one building from GitHub)
2. Go to the "Variables" tab
3. Add the following variables:

```env
ENV=production
FIREBASE_PROJECT_ID=your-project-id
FIREBASE_STORAGE_BUCKET=your-bucket-name.appspot.com
JWT_SECRET=your-secure-random-secret
MAX_UPLOAD_SIZE=52428800
```

4. **For Firebase Service Account:**

Instead of using a file path, we'll use JSON content:

- Open your `serviceAccountKey.json` file
- Copy the ENTIRE JSON content (including outer braces)
- Create a new variable: `FIREBASE_SERVICE_ACCOUNT_JSON`
- Paste the JSON as the value

Example:
```json
{
  "type": "service_account",
  "project_id": "trailmemo-prod",
  "private_key_id": "...",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
  ...
}
```

### Step 5: Run Database Migrations

Once PostgreSQL is provisioned:

1. Get your DATABASE_URL from Railway's Variables tab
2. Run migrations locally against the Railway database:

```bash
# Copy your DATABASE_URL from Railway
export DATABASE_URL="postgresql://postgres:..."

# Run migrations
psql $DATABASE_URL -f migrations/001_init.sql
```

Alternatively, you can run migrations from Railway's web terminal:
1. Click on PostgreSQL service
2. Click "Query"
3. Paste contents of `migrations/001_init.sql`
4. Execute

### Step 6: Deploy

Railway automatically deploys when you push to your repository:

```bash
git push origin main
```

Watch the build logs in Railway dashboard.

### Step 7: Get Your Production URL

1. In Railway dashboard, click on your service
2. Go to "Settings" → "Networking"
3. Click "Generate Domain"
4. Railway provides a URL like: `https://trailmemo-api-production.up.railway.app`

### Step 8: Test Production Deployment

```bash
# Test health endpoint
curl https://your-app.railway.app/health

# Should return:
# {"status":"ok","service":"trailmemo-api","version":"1.0.0"}
```

## Alternative: Docker Deployment

If you prefer Docker, the project includes a Dockerfile.

### Build and Run Locally

```bash
# Build image
docker build -t trailmemo-api .

# Run container
docker run -p 8080:8080 --env-file .env trailmemo-api
```

### Deploy to Any Docker Host

The Docker image can be deployed to:
- Google Cloud Run
- AWS ECS/Fargate
- Azure Container Instances
- DigitalOcean App Platform
- Heroku (Container Registry)
- Your own VPS with Docker

## Environment Variables Reference

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgresql://user:pass@host:5432/db` |
| `FIREBASE_PROJECT_ID` | Firebase project ID | `trailmemo-prod` |
| `FIREBASE_STORAGE_BUCKET` | Firebase storage bucket | `trailmemo-prod.appspot.com` |
| `FIREBASE_SERVICE_ACCOUNT_JSON` | Service account JSON content | `{"type":"service_account",...}` |
| `JWT_SECRET` | Secret for JWT signing | 32+ character random string |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `ENV` | Environment (development/production) | `development` |
| `MAX_UPLOAD_SIZE` | Max file size in bytes | `52428800` (50MB) |

## Post-Deployment Checklist

### Security

- [ ] All environment variables are set
- [ ] JWT secret is strong and unique
- [ ] Firebase Storage rules are configured
- [ ] CORS is configured for your domain
- [ ] HTTPS is enabled (automatic on Railway)
- [ ] Service account JSON is kept secure

### Firebase Configuration

1. **Set Storage Security Rules:**

Go to Firebase Console → Storage → Rules:

```javascript
rules_version = '2';
service firebase.storage {
  match /b/{bucket}/o {
    match /memos/{userId}/{fileName} {
      // Allow authenticated users to upload
      allow write: if request.auth != null && request.auth.uid == userId;
      // Allow anyone to read (for playback)
      allow read: if true;
    }
  }
}
```

2. **Set Firebase Auth settings:**

- Email/Password enabled
- Email verification (optional but recommended)
- Password policy configured

### Monitoring

1. **Railway Monitoring:**
   - Check "Metrics" tab for CPU/Memory usage
   - Review "Deployments" for build history
   - Check "Logs" for runtime errors

2. **Set up alerts:**
   - Railway can notify you of deployment failures
   - Consider adding external monitoring (UptimeRobot, etc.)

### Testing Production

Create a production test script:

```bash
#!/bin/bash
# test-production.sh

BASE_URL="https://your-app.railway.app"

echo "Testing health endpoint..."
curl -s $BASE_URL/health | jq

echo -e "\nTesting authentication (requires token)..."
# Add your test token
TOKEN="your_test_token"
curl -s $BASE_URL/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN" | jq
```

## Continuous Deployment

Railway automatically deploys on every push to your main branch.

### Deployment Workflow

```
git push origin main
    ↓
Railway detects push
    ↓
Builds Docker image
    ↓
Runs tests (if configured)
    ↓
Deploys new version
    ↓
Health checks pass
    ↓
Routes traffic to new version
```

### Rollback

If a deployment fails:

1. Go to Railway dashboard
2. Click on your service
3. Go to "Deployments" tab
4. Click on a previous successful deployment
5. Click "Redeploy"

## Scaling

### Railway Scaling

Railway automatically scales based on traffic, but you can adjust:

1. Go to Service Settings
2. Adjust "Memory" and "CPU" limits
3. Set "Replicas" (Pro plan required)

### Database Scaling

PostgreSQL on Railway:
- Free tier: Sufficient for MVP
- Pro tier: Larger databases, backups
- For heavy load: Consider managed PostgreSQL (AWS RDS, etc.)

## Monitoring and Logging

### View Logs

```bash
# Install Railway CLI
npm install -g @railway/cli

# Login
railway login

# View logs
railway logs
```

### Log Aggregation (Optional)

For production, consider:
- Logtail
- Papertrail  
- Datadog
- New Relic

Configure by adding log forwarding from Railway.

## Backup Strategy

### Database Backups

Railway Pro includes automatic backups. For free tier:

```bash
# Manual backup script
pg_dump $DATABASE_URL > backup-$(date +%Y%m%d).sql

# Restore from backup
psql $DATABASE_URL < backup-20241209.sql
```

### Automated Backups

Set up a cron job or GitHub Action to backup regularly.

## Cost Estimation

### Railway Pricing

**Hobby Plan (Free):**
- $5 credit per month
- PostgreSQL included
- Suitable for MVP and testing

**Developer Plan ($20/month):**
- Increased resources
- Database backups
- Better performance

**Typical Usage:**
- MVP with 10 users: Free tier sufficient
- 50+ active users: $20/month plan recommended
- 200+ users: $50-100/month depending on usage

### Firebase Costs

**Free Tier (Spark Plan):**
- 5GB Cloud Storage
- 1GB/day downloads
- 50K daily authentications

**Pay-as-you-go (Blaze Plan):**
- Only pay for what you use beyond free tier
- Storage: $0.026/GB/month
- Network: $0.12/GB

**Typical Usage:**
- 100 users, 50 memos/day: ~$0-5/month
- 500 users, 200 memos/day: ~$10-30/month

## Troubleshooting Production

### Common Issues

**1. Build Fails**
```bash
# Check Railway build logs
# Usually caused by:
- Missing dependencies in go.mod
- Build errors in code
- Insufficient memory during build
```

**2. Application Crashes**
```bash
# Check Railway logs
railway logs

# Common causes:
- Database connection failed (check DATABASE_URL)
- Firebase initialization failed (check service account JSON)
- Missing environment variables
```

**3. Database Connection Errors**
```bash
# Test database connection
psql $DATABASE_URL -c "SELECT version();"

# Ensure DATABASE_URL includes SSL mode
DATABASE_URL="postgresql://...?sslmode=require"
```

**4. File Upload Fails**
```bash
# Check Firebase Storage rules
# Verify FIREBASE_STORAGE_BUCKET is correct
# Test with smaller file first
```

## Best Practices

1. **Environment Management:**
   - Never commit secrets to git
   - Use different Firebase projects for dev/prod
   - Rotate JWT secrets periodically

2. **Database:**
   - Run migrations before deploying code changes
   - Keep migrations in version control
   - Test migrations on staging first

3. **Monitoring:**
   - Set up health check monitoring
   - Configure error alerting
   - Review logs regularly

4. **Security:**
   - Keep dependencies updated (`go get -u`)
   - Review Firebase security rules
   - Use strong JWT secrets
   - Enable HTTPS only

## Support

- **Railway Documentation:** https://docs.railway.app/
- **Firebase Documentation:** https://firebase.google.com/docs
- **Project Issues:** Open an issue on GitHub

---

🚀 **You're ready for production!**

Your TrailMemo API is now deployed and ready to serve your iOS app.

