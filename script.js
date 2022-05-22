import http from 'k6/http';
import exec from 'k6/execution';
import {check} from 'k6';

export let options = {
    vus: 5000,
    duration: "30s",
};

export default function () {
    let train = exec.vu.idInInstance % 100; // 100 different trains
	 
    let res = http.get('http://localhost:8080/book?train=' + train + '&person=1&seat=1');
    check(res, {
        'response code was 200': (res) => res.status == 200,
    });
}
