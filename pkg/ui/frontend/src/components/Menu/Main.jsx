import React from 'react';
import { Link } from 'react-router-dom';

import { makeStyles } from '@material-ui/styles';
import Paper from '@material-ui/core/Paper';
import Typography from '@material-ui/core/Typography';
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardHeader from '@material-ui/core/CardHeader';
import CardMedia from '@material-ui/core/CardMedia';
import CardContent from '@material-ui/core/CardContent';
import CardActions from '@material-ui/core/CardActions';
import CardActionArea from '@material-ui/core/CardActionArea';
import * as constants from '../../constants/constants';
import CssBaseline from '@material-ui/core/CssBaseline';

const useStyles = makeStyles({
    root: {
        margin: '0 auto',
        marginTop: 50,
        flexGrow: 1,
        // width: '50%',
        height: 400,
        textAlign: 'center',
    },
    item: {
        padding: '40px !important',
        textDecoration: 'none !important',
    },
    block: {
        backgroundColor: '#4169e1',
        height: '100%',
        width: '100%',
        padding: 40,
        // '&:hover': {
        //   backgroundColor: 'black',
        // },
    },
    link: {
        textDecoration: 'none',
        color: '#1890ff',
    },
});

const Main = props => {
    const classes = useStyles();

    return (
        <Paper className={classes.root}>
            <Typography variant={'h4'}>Welcome to Morphling</Typography>
            <br />
            <Grid
                className={classes.container}
                container
                spacing={16}
                direction="column"
                justify="center"
                alignItems="center"
            >
                <img
                    src={process.env.PUBLIC_URL + '/logo_menu.png'}
                    align="center"
                    height="60"
                    width="100"
                />

                <Grid item xs={0} className={classes.item} alignContent={'center'}>
                    <Paper className={classes.block}>
                        <Typography variant={'h6'} color={'secondary'}>
                            Automatic Configuration Recommendation
                        </Typography>
                        <Typography variant={'h6'} color={'secondary'}>
                            for Kubernetes Services
                        </Typography>
                    </Paper>
                </Grid>
            </Grid>
            <br />
        </Paper>
    );
};

export default Main;
