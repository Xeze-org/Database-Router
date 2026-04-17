package org.xeze.dbr;

import java.io.File;

public class Options {
    public String host;
    public boolean insecure;
    
    public File certFile;
    public File keyFile;
    public File caFile;

    public byte[] certData;
    public byte[] keyData;
    public byte[] caData;

    public Options(String host) {
        this.host = host;
    }
}
