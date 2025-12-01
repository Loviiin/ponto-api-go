import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '30s', target: 50 }, // Ramp up to 50 users over 30 seconds
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
        http_req_failed: ['rate<0.01'],   // Error rate must be less than 1%
    },
};

export default function () {
    // 1. Authenticate to get the token
    const loginUrl = 'http://localhost:8080/api/v1/auth/login';
    const loginPayload = JSON.stringify({
        email: 'colaborador@ponto.com',
        password: 'colabsenha',
    });
    const loginParams = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const loginRes = http.post(loginUrl, loginPayload, loginParams);

    check(loginRes, {
        'login successful': (r) => r.status === 200,
    });

    const token = loginRes.json('token');

    // 2. Perform the load test with the token
    const url = 'http://localhost:8080/api/v1/pontos'; // Corrected endpoint to plural
    const payload = JSON.stringify({
        // Add necessary payload fields here if required by the API
        // Example: { "tipo": "ENTRADA" } - Adjust based on actual API requirements
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
        },
    };

    const res = http.post(url, payload, params);

    check(res, {
        'status is 200 or 201': (r) => r.status === 200 || r.status === 201,
    });

    sleep(1);
}
