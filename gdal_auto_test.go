// Copyright 2014 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ingore
// +build ingore

package tiff

/*
// alg/gdalchecksum.cpp

// Compute checksum for image region.
//
// Computes a 16bit (0-65535) checksum from a region of raster data on a GDAL
// supported band.   Floating point data is converted to 32bit integer
// so decimal portions of such raster data will not affect the checksum.
// Real and Imaginary components of complex bands influence the result.
//
// @param hBand the raster band to read from.
// @param nXOff pixel offset of window to read.
// @param nYOff line offset of window to read.
// @param nXSize pixel size of window to read.
// @param nYSize line size of window to read.
//
// @return Checksum value.
//

int CPL_STDCALL
GDALChecksumImage( GDALRasterBandH hBand,
                   int nXOff, int nYOff, int nXSize, int nYSize )

{
    VALIDATE_POINTER1( hBand, "GDALChecksumImage", 0 );

    const static int anPrimes[11] =
        { 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43 };

    int  iLine, i, nChecksum = 0, iPrime = 0, nCount;
    GDALDataType eDataType = GDALGetRasterDataType( hBand );
    int  bComplex = GDALDataTypeIsComplex( eDataType );

    if (eDataType == GDT_Float32 || eDataType == GDT_Float64 ||
        eDataType == GDT_CFloat32 || eDataType == GDT_CFloat64)
    {
        double* padfLineData;
        GDALDataType eDstDataType = (bComplex) ? GDT_CFloat64 : GDT_Float64;

        padfLineData = (double *) VSIMalloc2(nXSize, sizeof(double) * 2);
        if (padfLineData == NULL)
        {
            CPLError( CE_Failure, CPLE_OutOfMemory,
                    "VSIMalloc2(): Out of memory in GDALChecksumImage. "
                    "Checksum value couldn't be computed\n");
            return 0;
        }

        for( iLine = nYOff; iLine < nYOff + nYSize; iLine++ )
        {
            if (GDALRasterIO( hBand, GF_Read, nXOff, iLine, nXSize, 1,
                              padfLineData, nXSize, 1, eDstDataType, 0, 0 ) != CE_None)
            {
                CPLError( CE_Failure, CPLE_FileIO,
                        "Checksum value couldn't be computed due to I/O read error.\n");
                break;
            }
            nCount = (bComplex) ? nXSize * 2 : nXSize;

            for( i = 0; i < nCount; i++ )
            {
                double dfVal = padfLineData[i];
                int nVal;
                if (CPLIsNan(dfVal) || CPLIsInf(dfVal))
                {
                    // Most compilers seem to cast NaN or Inf to 0x80000000.
                    // but VC7 is an exception. So we force the result
                    // of such a cast
                    nVal = 0x80000000;
                }
                else
                {
                    // Standard behaviour of GDALCopyWords when converting
                    // from floating point to Int32
                    dfVal += 0.5;

                    if( dfVal < -2147483647.0 )
                        nVal = -2147483647;
                    else if( dfVal > 2147483647 )
                        nVal = 2147483647;
                    else
                        nVal = (GInt32) floor(dfVal);
                }

                nChecksum += (nVal % anPrimes[iPrime++]);
                if( iPrime > 10 )
                    iPrime = 0;

                nChecksum &= 0xffff;
            }
        }

        CPLFree(padfLineData);
    }
    else
    {
        int  *panLineData;
        GDALDataType eDstDataType = (bComplex) ? GDT_CInt32 : GDT_Int32;

        panLineData = (GInt32 *) VSIMalloc2(nXSize, sizeof(GInt32) * 2);
        if (panLineData == NULL)
        {
            CPLError( CE_Failure, CPLE_OutOfMemory,
                    "VSIMalloc2(): Out of memory in GDALChecksumImage. "
                    "Checksum value couldn't be computed\n");
            return 0;
        }

        for( iLine = nYOff; iLine < nYOff + nYSize; iLine++ )
        {
            if (GDALRasterIO( hBand, GF_Read, nXOff, iLine, nXSize, 1,
                            panLineData, nXSize, 1, eDstDataType, 0, 0 ) != CE_None)
            {
                CPLError( CE_Failure, CPLE_FileIO,
                        "Checksum value couldn't be computed due to I/O read error.\n");
                break;
            }
            nCount = (bComplex) ? nXSize * 2 : nXSize;

            for( i = 0; i < nCount; i++ )
            {
                nChecksum += (panLineData[i] % anPrimes[iPrime++]);
                if( iPrime > 10 )
                    iPrime = 0;

                nChecksum &= 0xffff;
            }
        }

        CPLFree( panLineData );
    }

    return nChecksum;
}
*/
