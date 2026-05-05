import { Wallets } from 'fabric-network';
import FabricCAServices  from 'fabric-ca-client';
import path from 'path';

export async function enrollUser(userId: string, password: string) {
    // Use Bevel operator CA URL and helper to enroll
    const caURL = process.env.CA_URL!;
    const walletPath = path.join(process.cwd(), 'wallet');
    const wallet = await Wallets.newFileSystemWallet(walletPath);

    const caClient = new FabricCAServices(caURL);
    const enrollment = await caClient.enroll({ enrollmentID: userId, enrollmentSecret: password });
    const identity = {
        credentials: {
            certificate: enrollment.certificate,
            privateKey: enrollment.key.toBytes(),
        },
        mspId: 'Org1MSP',
        type: 'X.509',
    };
    await wallet.put(userId, identity);
}