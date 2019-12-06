import {BN256G1Point} from '@dedis/kyber/pairing/point';
import {KeyPair, LibDrynx, Suite} from "../src";

describe('pairing', function () {
    it('should add and subtract g1 points', () => {
        const p1 = new BN256G1Point(1);
        const p2 = new BN256G1Point(2);

        const aa = new BN256G1Point().add(p1, p2);
        const bb = new BN256G1Point().sub(p1, p2.clone().neg(p2));

        expect(aa.equals(bb)).toBeTruthy();
    });

    it('should negate points', () =>{
        const s = Suite.scalar().one();
        const one = Suite.base();
        expect(one.equals(Suite.point().mul(s, Suite.base())));
        s.add(s, Suite.scalar().one());
        one.add(one, Suite.base());
        expect(one.equals(Suite.point().mul(s, Suite.base())));

        s.neg(s);
        one.neg(one);
        expect(one.equals(Suite.point().mul(s, Suite.base())));
    });

    it('should create keypair', () => {
        const kp1 = KeyPair.random();
        const kp2 = KeyPair.random();
        expect(kp1.scalar.equals(kp2.scalar)).toBeFalse();
        expect(kp1.point.equals(kp2.point)).toBeFalse();

        const s = Suite.scalar().add(kp1.scalar, kp2.scalar);
        const p = Suite.point().add(kp1.point, kp2.point);
        const pp = Suite.point().mul(s, Suite.base());
        expect(pp.equals(p)).toBeTruthy();
    });

    it('should solve the discreet log problem for positive numbers', () => {
        const mi = LibDrynx.maxInt;
        for (let ii of [0, 1, mi / 2, mi / 3, mi + 1]) {
            const i = Math.floor(ii);
            const enc = LibDrynx.intToPoint(i);
            if (i <= mi) {
                const dec = LibDrynx.pointtoInt(enc, false);
                expect(i).toBe(dec);
            } else {
                expect(() => {
                    LibDrynx.pointtoInt(enc, false)
                }).toThrow();
            }
        }
    });

    it('should solve the discreet log problem for negative numbers', () => {
        const mi = LibDrynx.maxInt;
        for (let ii of [0, -1, -mi / 2, -mi / 3, -mi - 1]) {
            const i = Math.floor(ii);
            const enc = LibDrynx.intToPoint(i);
            if (Math.abs(i) <= mi) {
                const dec = LibDrynx.pointtoInt(enc, true);
                expect(i).toBe(dec);
            } else {
                expect(() => {
                    LibDrynx.pointtoInt(enc, true)
                }).toThrow();
            }
        }
    });

    it('should crypt and decrypt a point', () => {
        const kp = KeyPair.random();
        const p = KeyPair.random().point;
        const enc = LibDrynx.encryptPoint(kp.point, p);
        const dec = LibDrynx.decryptPoint(kp.scalar, enc.CT);
        expect(dec.equals(p)).toBeTrue();
    });

    it('should crypt and decrypt an integer', () => {
        const kp = KeyPair.random();
        const mi = LibDrynx.maxInt;
        for (let ii of [0, 1, mi / 2, mi / 3, mi + 1]) {
            const iint = Math.floor(ii);
            for (let i of [iint, -iint]) {
                const enc = LibDrynx.encryptInt(kp.point, i);
                if (Math.abs(i) <= mi) {
                    const dec = LibDrynx.decryptInt(kp.scalar, enc, true);
                    expect(dec).toBe(i);
                } else {
                    expect(() => {
                        LibDrynx.decryptInt(kp.scalar, enc, true)
                    }).toThrow();
                }
            }
        }
    })
});
