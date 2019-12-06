import {BN256G1Point} from '@dedis/kyber/pairing/point';
import BN256Scalar from '@dedis/kyber/pairing/scalar';
// @ts-ignore
import {randomBytes} from "crypto-browserify";
import {Point, Scalar} from "@dedis/kyber";

/**
 * Have a central definition of the points used which creates the desired points
 * and scalars. Similar to libunlynx.SuiTe
 */
export class Suite {
    static point(): Point {
        return new BN256G1Point();
    }

    static scalar(): Scalar {
        return new BN256Scalar();
    }

    static base(): Point {
        return this.point().base();
    }
}

/**
 * Holds a keypair of a scalar/point pair, where the point is the scalar * Base,
 * which corresponds most of the time to a private/public keypair.
 */
export class KeyPair {
    point: Point;

    constructor(public scalar: Scalar) {
        this.point = new BN256G1Point();
        const base = new BN256G1Point().base();
        this.point.mul(scalar, base);
    }

    static random(): KeyPair {
        // TODO: find out why this fails to pass the "Should create keypair" test if it is 32!?!
        const s = new BN256Scalar(randomBytes(31));
        return new KeyPair(s);
    }
}

/**
 * Represents an ElGamal encrypted CipherText.
 */
interface CipherText {
    K: Point,
    C: Point,
}

/**
 * Return from the encrypted method including the random factor.
 */
interface Encrypted {
    CT: CipherText;
    r: Scalar;
}

/**
 * Basic LibDrynx methods that can be used to encrypt and decrypt values.
 */
export class LibDrynx {
    static maxInt = 100;

    /**
     * Encrypts an integer so that it can be used in homomorphic operations.
     * @param pub
     * @param i
     */
    static encryptInt(pub: Point, i: number): CipherText {
        const enc = this.encryptPoint(pub, this.intToPoint(i));
        return enc.CT;
    }

    /**
     * Decrypts the integer found in the point. As the final step of the decryption
     * involves solving the discreet log problem, it will only solve for integers < maxInt.
     * @param priv
     * @param cipher
     * @param checkNeg
     */
    static decryptInt(priv: Scalar, cipher: CipherText, checkNeg: boolean = false): number {
        const M = this.decryptPoint(priv, cipher);
        return this.pointtoInt(M, checkNeg);
    }

    /**
     * ElGamal encryption of a message that has been mapped to a point m.
     * @param pub
     * @param m
     */
    static encryptPoint(pub: Point, m: Point): Encrypted {
        const B = Suite.base();
        const r = Suite.scalar().pick((l: number) => randomBytes(l));
        const K = Suite.point().mul(r, B);
        const S = Suite.point().mul(r, pub);
        const C = Suite.point().add(S, m);
        return {CT: {K, C}, r};
    }

    /**
     * ElGamal decryption.
     * @param priv
     * @param c
     */
    static decryptPoint(priv: Scalar, c: CipherText): Point {
        const S = Suite.point().mul(priv, c.K);
        return Suite.point().sub(c.C, S);
    }

    /**
     * Paillier encryption (at least I think so ;) of a number to let it be used in
     * homomorphic operations.
     * @param n
     */
    static intToPoint(n: number): Point {
        const neg = n < 0;
        if (neg){
            n = n * -1;
        }
        const i64 = Buffer.alloc(8);
        i64.writeInt32BE(n, 4);
        const i = Suite.scalar().setBytes(Buffer.from(i64));
        if (neg){
            i.neg(i);
        }
        return Suite.point().mul(i, Suite.base());
    }

    /**
     * This solves the discreet log problem by simply trying out all scalar values in
     * ascending value.
     * If checkNeg is true, it tries to find the number in a zig-zag way: 0, 1, -1, 2, -2, ...
     * TODO: add a static table to speed up lookups of known numbers.
     * TODO: to make it faster, one could directly work with points:
     * p1 = 1 * base
     * p2 = 2 * base = p1 + base
     * -> so instead of doing the mul for every step, which is very expensive, one could
     * simply add the base at every step.
     * I tried it out and found some strange things happening that should not be happening.
     * So for the time being this is a future optimization...
     * @param p
     * @param checkNeg
     */
    static pointtoInt(p: Point, checkNeg: boolean): number {
        const b = Suite.base();
        if (p.equals(Suite.point().mul(Suite.scalar().zero(), b))) {
            return 0;
        }
        const one = Suite.scalar().one();
        const s = Suite.scalar().one();
        let i = 1;
        while (true) {
            if (p.equals(Suite.point().mul(s, b))) {
                return i;
            }
            if (checkNeg) {
                i = i * -1;
                s.neg(s);
                if (i < 0) {
                    continue;
                }
            }
            i = i + 1;
            s.add(s, one);
            if (i > this.maxInt){
                throw new Error("didn't find discreet log");
            }
        }
    }
}
