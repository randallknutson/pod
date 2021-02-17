
import configparser
from collections import namedtuple
from Crypto.Cipher import AES
from Crypto.Hash import CMAC
from CryptoMobile.Milenage import Milenage
import curve25519
from textwrap import wrap
import sys
import argparse

# noqa: E501


# https://tools.ietf.org/html/rfc3748
CODE_BIN = {
    1: 'Request',
    2: 'Response',
    3: 'Success',
    4: 'Failure'

}

SUBTYPE = {
    1: "AKA-Challenge",
    2: "AKA-Authentication-Reject",
    4: "AKA-Synchronization-Failure",
    5: "AKA-Identity",
    10: "SIM-Start",
    11: "SIM-Challenge",
    12: "AKA-Notification and SIM-Notification",
    13: "AKA-Reauthentication and SIM-Reauthentication",
    14: "AKA-Client-Error and SIM-Client-Error",
}

ATTRIBUTE_TYPE = {
    1: "AT_RAND",
    2: "AT_AUTN",
    3: "AT_RES",
    4: "AT_AUTS",
    6: "AT_PADDING",
    7: "AT_NONCE_MT",
    10: "AT_PERMANENT_ID_REQ",
    11: "AT_MAC",
    12: "AT_NOTIFICATION",
    13: "AT_ANY_ID_REQ",
    14: "AT_IDENTITY",
    15: "AT_VERSION_LIST",
    16: "AT_SELECTED_VERSION",
    17: "AT_FULLAUTH_ID_REQ",
    19: "AT_COUNTER",
    20: "AT_COUNTER_TOO_SMALL",
    21: "AT_NONCE_S",
    22: "AT_CLIENT_ERROR_CODE",
    126: "???_CONTROLER_OR_NODE_IV??",
    129: "AT_IV",
    130: "AT_ENCR_DATA",
    132: "AT_NEXT_PSEUDONYM",
    133: "AT_NEXT_REAUTH_ID",
    134: "AT_CHECKCODE",
    135: "AT_RESULT_IND",
}


class Ble():
    def __init__(self, input=None, bin=None, eap=False):
        if not bin:
            hex = input.split(',')
            bin = list(map(lambda x: int(x, 16), hex))
        self._from = bin[8:12]
        self._to = bin[12:16]
        self.preamble = bin[0]
        self.unknown = bin[1:7]
        self.crc = bin[-3:]
        self.data = bin[16:]
        self.eap = None
        if eap:
            self.eap = Eap(bin=self.data)

    def __str__(self):
        ret = f"""
            Preamble: {self.preamble}
            From: {self._from.hex()}
            To: {self._to.hex()}
            Data: {self.data} :: {len(self.data)}
            Data Hex: {self.data.hex()}
            Unknown: {self.unknown}
            UnknownHex: {self.unknown.hex()}
        """
        if self.eap:
            ret += str(self.eap)
        return ret


class Attribute():
    def __init__(self, type, data):
        self.type_bin = type
        self.type = ATTRIBUTE_TYPE.get(self.type_bin, "not documented")
        self.data = data

    def __str__(self):
        return f"""
                Type: {self.type} :: {self.type_bin}
                Data: {self.data} :: {len(self.data)}
                Hex: {self.data.hex()}
        """


class Res(Attribute):
    def __init__(self, data):
        self.type_bin = 3
        self.type = ATTRIBUTE_TYPE.get(self.type_bin, "not documented")
        self.length = (data[0] * 16 + data[1]) / 8
        self.data = data[2:]

    def __str__(self):
        return f"""
                Type: {self.type} :: {self.type_bin}
                Data: {self.data}
                Length: {self.length}
                Hex: {self.data.hex()}
        """


DECODED_ATTRIBUTES = {
    3: Res,
}


class Eap():

    def __init__(self, input=None, bin=None):
        """
        Example input
            01,48,00,38,17,01,00,00,02,05,00,00,5d,2a,6e,39,dc,59,b9,b9,04,cd,89,cc,42,3c,c1,53,01,05,00,00,89,ce,f1,27,62,ee,df,fe,25,d6,41,90,b7,16,50,c9,7e,02,00,00,cb,07,56,3b,
        """
        if not bin:
            hex = input.split(',')
            bin = list(map(lambda x: int(x, 16), hex))

        self.length = bin[2] * 16 + bin[3]
        self.code_bin = bin[0]
        self.code = CODE_BIN[self.code_bin]
        self.identifier = bin[1]
        self.type = None
        if self.length <= 4:
            return
        self.type = str(bin[4]) + "(should be 23)"
        self.subtype_int = bin[5]
        self.subtype = SUBTYPE[self.subtype_int]
        self.attributes = []
        tail = bin[8:]
        while tail:
            attribute_type = tail[0]
            length = tail[1] * 4  # multiple of 4 bytes
            attribute_data = tail[2:length]

            if attribute_type in [1, 2]:  # remove 2 reserved bytes
                assert(attribute_data[0] == 0)
                assert(attribute_data[1] == 0)

                attribute_data = attribute_data[2:]

            attribute = Attribute(attribute_type, attribute_data)
            if attribute_type in DECODED_ATTRIBUTES:
                attribute = DECODED_ATTRIBUTES[attribute_type](attribute_data)

            self.attributes.append(attribute)
            tail = tail[length:]

    def __str__(self):
        ret = f"""
        Length: {self.length}
        Code:: {self.code}
        Identifier: {self.identifier}
        """
        if self.type is not None:
            ret += f"""
                Type: {self.type}
                Subtype: {self.subtype} :: {self.subtype_int}
                Attributes: {len(self.attributes)}
            """
            for i in self.attributes:
                ret += str(i)
        return ret


class PodCommand():
    def __init__(self, cmd):
        self.cmd = cmd
        self.data = list(map(lambda x: int(x, 16), wrap(cmd, 2)))

    def __str__(self):
        return f"""
            Raw: {self.cmd}
            Data: {self.data}
            Hex: {self.data.hex()}
        """


def decrypt(args):
    fields = [
        "packet_data",
        "expected",
        "nonce",
        "ck",
        "tag",
        "header",
    ]
    if args.list_fields:
        print(fields)
        sys.exit()
    data = config_to_input("decrypt", args.config, fields)

    print(f"Decrypt: CK: {data.ck.hex()}, Nonce: {data.nonce.hex()}, Data:{data.packet_data.hex()}")

    actualData = data.packet_data
    
    print(f"Actual data: {actualData.hex()} :: {len(actualData)}")
    print(f"Header: {data.header.hex()} :: {len(data.header)}")

    tag = data.tag
    print(f"Tag: {tag.hex()} :: {len(tag)}")

    cipher = AES.new(data.ck, AES.MODE_CCM, data.nonce, mac_len=8)
    cipher.update(data.header)
    decrypted = cipher.decrypt(actualData)
    
    try:    
        print("Verify: ", cipher.verify(tag))
        print("VERIFIED!")
    except ValueError:
        print("NOPE")
    print("Decrypted: ", decrypted.hex(), len(decrypted))
    print("Expected:  ", data.expected.hex(), len(data.expected))
    
    cipher = AES.new(data.ck, AES.MODE_CCM, data.nonce, mac_len=8)
    cipher.update(data.header)
    encrypted, digest = cipher.encrypt_and_digest(decrypted)
    print("Encrypted: ", encrypted.hex(), len(encrypted))
    print("Digest: ", digest.hex(), len(digest))


def encrypt(args):
    fields = [
        "packet_data",
        "command",
        "nonce",
        "ck"
    ]
    if args.list_fields:
        print(fields)
        sys.exit()
    data = config_to_input("encrypt", args.config, fields)
    ble = Ble(bin=data.packet_data)
    print(f"Encrypt: CK: {data.ck.hex()}, Nonce: {data.nonce.hex()}, Data:{data.packet_data.hex()}")

    cipher = AES.new(data.ck, AES.MODE_CCM, data.nonce)
    encrypted = cipher.encrypt(data.command)

    print("Encrypted: ", encrypted.hex(), len(encrypted))
    print("Expected:  ", ble.data.hex(), len(ble.data))


def milenage(bin_ltk, bin_rand, bin_seq, bin_ck, want_res, bin_autn):
    print("LTK: ", bin_ltk.hex(), len(bin_ltk))
    print("Autn: ", bin_autn.hex(","), len(bin_autn))
    print("Rand: ", bin_rand.hex(), len(bin_rand))
    print("Seq: ", bin_seq.hex(","), len(bin_seq))
    bin_amf = bin_autn[6:8]
    print("Amf: ", bin_amf.hex(), len(bin_amf))
    # 0xf6203e12d502c2cd
    # 0x18b32cc76a676d2b

    op = 'f6203e12d502c2cd18b32cc76a676d2b'
    # op = '18b32cc76a676d2bf6203e12d502c2cd'
    op = 'cdc202d5123e20f62b6d676ac72cb318'
    bin_op = bytes.fromhex(op)
    print("Op: ", bin_op.hex(","), len(bin_op))

    m = Milenage(bin_op)
    res, res_ck, ik, ak = m.f2345(bin_ltk, bin_rand)

    print("CK is: ", bin_ck.hex(","), len(bin_ck))
    print("Received CK is: ", res_ck.hex(","), len(res_ck))
    print("Got  Res is: ", res.hex(","), len(res))
    print("Want Res is: ", want_res.hex(","), len(want_res))

    print("IK is: ", ik.hex(","), len(ik))
    print("AK is: ", ak.hex(","), len(ak))


def _xor(a, b):
    ret = bytearray(16)
    for i in range(len(a)):
        ret[i] = a[i] ^ b[i]
    return ret


def _get_milenage(opc, k, rand, sqn, amf):
    '''
        Computes milenage values from OPc, K, RAND, SQN and AMF
        Returns a concatenated list (RAND + AUTN + IK + CK + RES) that
        will be sent back as the response to the client (hostapd). This
        is a python re-write of the function eap_aka_get_milenage() from
        src/simutil.c
    '''
    # rand gets returned, so it should be left as a hex string
    _rand = rand

    aes1 = AES.new(bytes(k), AES.MODE_ECB)
    tmp1 = _xor(_rand, opc)
    tmp1 = aes1.encrypt(bytes(tmp1))
    tmp1 = bytearray(tmp1)
    tmp2 = bytearray()
    tmp2[0:6] = sqn
    tmp2[6:2] = amf
    tmp2[9:6] = sqn
    tmp2[15:2] = amf
    tmp3 = bytearray(16)
    for i in range(len(tmp1)):
        tmp3[(i + 8) % 16] = tmp2[i] ^ opc[i]
    tmp3 = _xor(tmp3, tmp1)
    aes2 = AES.new(bytes(k), AES.MODE_ECB)
    tmp1 = aes2.encrypt(bytes(tmp3))
    tmp1 = bytearray(tmp1)
    tmp1 = _xor(tmp1, opc)
    maca = _bytetostring(tmp1[0:8])
    tmp1 = _xor(_rand, opc)
    aes3 = AES.new(bytes(k), AES.MODE_ECB)
    tmp2 = aes3.encrypt(bytes(tmp1))
    tmp2 = bytearray(tmp2)
    tmp1 = _xor(tmp2, opc)
    tmp1[15] ^= 1
    aes4 = AES.new(bytes(k), AES.MODE_ECB)
    tmp3 = aes4.encrypt(bytes(tmp1))
    tmp3 = bytearray(tmp3)
    tmp3 = _xor(tmp3, opc)
    res = _bytetostring(tmp3[8:16])
    ak = _bytetostring(tmp3[0:6])
    for i in range(len(tmp1)):
        tmp1[(i + 12) % 16] = tmp2[i] ^ opc[i]
    tmp1[15] ^= 1 << 1
    aes5 = AES.new(bytes(k), AES.MODE_ECB)
    tmp1 = aes5.encrypt(bytes(tmp1))
    tmp1 = bytearray(tmp1)
    tmp1 = _xor(tmp1, opc)
    ck = _bytetostring(tmp1)
    for i in range(len(tmp1)):
        tmp1[(i + 8) % 16] = tmp2[i] ^ opc[i]
    tmp1[15] ^= 1 << 2
    aes6 = AES.new(bytes(k), AES.MODE_ECB)
    tmp1 = aes6.encrypt(bytes(tmp1))
    tmp1 = bytearray(tmp1)
    tmp1 = _xor(tmp1, opc)
    ik = _bytetostring(tmp1)
    tmp1 = bytearray.fromhex(ak)
    autn = bytearray(6)
    for i in range(0, 6):
        autn[i] = sqn[i] ^ tmp1[i]
    autn[6:2] = amf
    autn[8:8] = bytearray.fromhex(maca)[0:8]
    autn = _bytetostring(autn)
    return rand, autn, ik, ck, res


def _bytetostring(b):
    return ''.join(format(x, '02x') for x in b)


def cmac(key, data):
    aes = CMAC.new(key, ciphermod=AES)
    mac = aes.update(data).digest()
    # print("mac ", mac.hex(), " data len", len(data))
    return mac


def config_to_input(section, config, fields):
    Input = namedtuple("Input", fields)
    bin_data = dict()
    for f in fields:
        field_value = config.get(section, f).replace(",", "")
        bin_data[f] = bytes.fromhex(field_value)
    return Input(**bin_data)


def key_exchange(args):
    fields = [
        "pdm_ltk",
        "pdm_public",
        "pdm_nonce",
        "pdm_sps2",

        "pod_nonce",
        "pod_secret",
        "pod_public",
        "pod_ltk",
        "pod_sps2",
    ]

    if args.list_fields:
        print(fields)
        sys.exit()
    data = config_to_input("key_exchange", args.config, fields)

    private = curve25519.Private()
    private.private = data.pod_secret
    public = private.get_public()
    print(public.serialize().hex())
    pdm_curve_public = curve25519.Public(data.pdm_public)
    shared_secret = private.get_shared_key(pdm_curve_public, hashfunc=lambda x: x)
    print("POD LTK:                ", data.pod_ltk.hex())
    print("Donna LTK double check: ", shared_secret.hex(), "\n")

    first_key = data.pod_public[-4:] + data.pdm_public[-4:] + data.pod_nonce[-4:] + data.pdm_nonce[-4:]
    temp_mac = cmac(first_key, shared_secret)

    bb_data = bytes.fromhex("01") + bytes("TWIt", "ascii") + data.pod_nonce + data.pdm_nonce + bytes.fromhex("0001")
    bb = cmac(temp_mac, bb_data)

    ab_data = bytes.fromhex("02") + bytes("TWIt", "ascii") + data.pod_nonce + data.pdm_nonce + bytes.fromhex("0001")
    ab = cmac(temp_mac, ab_data)
    print("LTK :     ", ab.hex())
    print("Expected: ", data.pdm_ltk.hex(), "\n")

    pdm_conf_data = bytes("KC_2_U", "ascii") + data.pdm_nonce + data.pod_nonce
    pdm_conf = cmac(bb, pdm_conf_data)
    print("PDM conf: ", pdm_conf.hex())
    print("Expected:  ", data.pdm_sps2.hex(), "\n")

    pod_conf_data = bytes("KC_2_V", "ascii") + data.pod_nonce + data.pdm_nonce  # ???
    pod_conf = cmac(bb, pod_conf_data)
    print("PDM conf:  ", pod_conf.hex())
    print("Expected:  ", data.pod_sps2.hex())


def eap_aka(args):
    fields = [
        "ck",
        "rand",
        "seq",
        "ltk",
        "res",
        "autn",
    ]

    if args.list_fields:
        print(fields)
        sys.exit()
    data = config_to_input("eap_aka", args.config, fields)

    milenage(data.ltk, data.rand, data.seq, data.ck, data.res, data.autn)
    # print("-----------")
    # this is not working, because _get_milenage expects OPc instead of OP
    # amf = "b9b9"
    # opc = "cdc202d5123e20f62b6d676ac72cb318"
    # rand, autn, ik, ck, res = _get_milenage(bytes.fromhex(opc), data.ltk, data.rand, data.seq, bytes.fromhex(amf))

    # print("Rand: ", rand.hex(","))
    # print("Autn: ", bytes.fromhex(autn).hex(","))
    # print("CK: ", bytes.fromhex(ck).hex(","))
    # print("Res: ", bytes.fromhex(res).hex(","))
    # print("Ik: ", bytes.fromhex(ik).hex(","))


def eap_print(args):
    fields = [
        "packet_from_log"
    ]
    if args.list_fields:
        print(fields)
        sys.exit()
    data = config_to_input("eap_print", args.config, fields)
    print(Eap(bin=data.packet_from_log))


functions = dict(
    key_exchange=key_exchange,
    eap_aka=eap_aka,
    decrypt=decrypt,
    encrypt=encrypt,
    eap_print=eap_print,
)

if __name__ == "__main__":
    parser = argparse.ArgumentParser("Encryption experiments")
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument("-i", "--input", help="File containing test data")
    group.add_argument("-l", "--list-fields", action="store_true", help="List the expected fields in the test data file for this experiment")
    subparsers = parser.add_subparsers()
    parser.set_defaults(func=None)
    for name, function in functions.items():
        subparser = subparsers.add_parser(name)
        subparser.set_defaults(func=function)

    args = parser.parse_args()
    if args.input:
        args.config = configparser.ConfigParser()
        args.config.read(args.input)

    if not args.func:
        parser.print_help()
        sys.exit()
    args.func(args)
